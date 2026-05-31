#!/usr/bin/env bash
set -euo pipefail

K8S_NAMESPACE="${K8S_NAMESPACE:-go-svc}"
PROM_PORT="${OBS_PROM_PORT:-19090}"
JAEGER_PORT="${OBS_JAEGER_PORT:-16686}"
PASS=0
FAIL=0
PF_PROM=""
PF_JAEGER=""

cleanup() {
  if [ -n "${PF_PROM:-}" ]; then kill "$PF_PROM" 2>/dev/null || true; wait "$PF_PROM" 2>/dev/null || true; fi
  if [ -n "${PF_JAEGER:-}" ]; then kill "$PF_JAEGER" 2>/dev/null || true; wait "$PF_JAEGER" 2>/dev/null || true; fi
}
trap cleanup EXIT

echo "=== K8s Observability Verification ==="

echo "--- Port-forward ---"
kubectl port-forward -n "$K8S_NAMESPACE" svc/prometheus "$PROM_PORT:9090" &
PF_PROM=$!
kubectl port-forward -n "$K8S_NAMESPACE" svc/jaeger "$JAEGER_PORT:16686" &
PF_JAEGER=$!
sleep 1

# Verify port-forward processes are alive
for pid_var in PF_PROM PF_JAEGER; do
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

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
