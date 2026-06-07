#!/usr/bin/env bash
set -euo pipefail

K8S_NAMESPACE="${K8S_NAMESPACE:-go-svc}"
PROM_PORT="${OBS_PROM_PORT:-19090}"
JAEGER_PORT="${OBS_JAEGER_PORT:-16686}"
LOKI_PORT="${OBS_LOKI_PORT:-13100}"
GATEWAY_PORT="${OBS_GATEWAY_PORT:-18080}"
PASS=0
FAIL=0
PF_PROM=""
PF_JAEGER=""
PF_LOKI=""
PF_GATEWAY=""

cleanup() {
  if [ -n "${PF_PROM:-}" ]; then kill "$PF_PROM" 2>/dev/null || true; wait "$PF_PROM" 2>/dev/null || true; fi
  if [ -n "${PF_JAEGER:-}" ]; then kill "$PF_JAEGER" 2>/dev/null || true; wait "$PF_JAEGER" 2>/dev/null || true; fi
  if [ -n "${PF_LOKI:-}" ]; then kill "$PF_LOKI" 2>/dev/null || true; wait "$PF_LOKI" 2>/dev/null || true; fi
  if [ -n "${PF_GATEWAY:-}" ]; then kill "$PF_GATEWAY" 2>/dev/null || true; wait "$PF_GATEWAY" 2>/dev/null || true; fi
}
trap cleanup EXIT

echo "=== K8s Observability Verification ==="

echo "--- Port-forward ---"
kubectl port-forward -n "$K8S_NAMESPACE" svc/prometheus "$PROM_PORT:9090" &
PF_PROM=$!
kubectl port-forward -n "$K8S_NAMESPACE" svc/jaeger "$JAEGER_PORT:16686" &
PF_JAEGER=$!
kubectl port-forward -n "$K8S_NAMESPACE" svc/loki "$LOKI_PORT:3100" &
PF_LOKI=$!
kubectl port-forward -n "$K8S_NAMESPACE" svc/gateway-api "$GATEWAY_PORT:8080" &
PF_GATEWAY=$!
sleep 2

for pid_var in PF_PROM PF_JAEGER PF_LOKI PF_GATEWAY; do
  eval "pid=\$$pid_var"
  if ! kill -0 "$pid" 2>/dev/null; then
    echo "  FAIL: port-forward ($pid_var=$pid) died immediately (port in use?)"
    exit 1
  fi
done

echo "--- Prometheus ---"
for svc in gateway-api user-rpc product-rpc inventory-rpc order-rpc; do
  result=$(curl -sf "http://localhost:$PROM_PORT/api/v1/query?query=up%7Binstance%3D%22${svc}:6060%22%7D" 2>/dev/null || echo "")
  if echo "$result" | grep -q '"value":\[.*,"1"\]'; then
    echo "  PASS: $svc:6060 target UP"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: $svc:6060 target not UP"
    FAIL=$((FAIL + 1))
  fi
done

echo "--- Jaeger ---"
JAEGER=$(curl -sf "http://localhost:$JAEGER_PORT/api/services" 2>/dev/null | grep -c '"data"' || echo "0")
JAEGER=$(echo "$JAEGER" | tr -d '[:space:]')
if [ "${JAEGER:-0}" -gt 0 ]; then
  echo "  PASS: Jaeger services found"
  PASS=$((PASS + 1))
else
  echo "  FAIL: Jaeger not responding"
  FAIL=$((FAIL + 1))
fi

echo "--- Loki ---"
LOKI_READY=0
for _ in $(seq 1 10); do
  if curl -sf "http://localhost:$LOKI_PORT/ready" 2>/dev/null | grep -qi "ready"; then
    LOKI_READY=1
    break
  fi
  sleep 2
done
if [ "$LOKI_READY" -eq 1 ]; then
  echo "  PASS: Loki ready"
  PASS=$((PASS + 1))
else
  echo "  FAIL: Loki not ready"
  FAIL=$((FAIL + 1))
fi

echo "--- Loki Logs ---"
SUFFIX="$(date +%s)-$$"
REGISTER_HEADERS=$(curl -sD - -o /dev/null \
  -X POST "http://localhost:$GATEWAY_PORT/api/v1/register" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"trace-k8s-${SUFFIX}\",\"password\":\"123456\"}" 2>/dev/null || echo "")
TRACE_ID=$(echo "$REGISTER_HEADERS" | grep -i "X-Trace-Id:" | awk '{print $2}' | tr -d '\r')
sleep 5

LOGS=$(curl -sf --data-urlencode 'query={namespace="go-svc"}' "http://localhost:$LOKI_PORT/loki/api/v1/query_range" 2>/dev/null || echo "")
LOG_COUNT=$(echo "$LOGS" | grep -c '"stream"' || echo "0")
LOG_COUNT=$(echo "$LOG_COUNT" | tr -d '[:space:]')
if [ "${LOG_COUNT:-0}" -gt 0 ]; then
  echo "  PASS: logs found via LogQL ($LOG_COUNT streams)"
  PASS=$((PASS + 1))
else
  echo "  FAIL: no logs returned by LogQL (Alloy may not be scraping yet)"
  FAIL=$((FAIL + 1))
fi

echo "--- Trace-Log Correlation ---"
if [ -n "${TRACE_ID:-}" ]; then
  MATCHES=$(curl -sf --data-urlencode "query={namespace=\"go-svc\", app=\"user-rpc\"} | json | trace_id=\"$TRACE_ID\"" \
    "http://localhost:$LOKI_PORT/loki/api/v1/query_range" 2>/dev/null | grep -c '"stream"' || echo "0")
  MATCHES=$(echo "$MATCHES" | tr -d '[:space:]')
  if [ "${MATCHES:-0}" -gt 0 ]; then
    echo "  PASS: log found by trace_id=$TRACE_ID ($MATCHES)"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: no log for trace_id=$TRACE_ID in user-rpc"
    FAIL=$((FAIL + 1))
  fi
else
  echo "  FAIL: X-Trace-Id header missing from register response"
  FAIL=$((FAIL + 1))
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
