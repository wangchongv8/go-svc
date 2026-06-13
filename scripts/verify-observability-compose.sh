#!/usr/bin/env bash
set -euo pipefail

BASE="${BASE_URL:-http://localhost:8080}"
PASS=0
FAIL=0

cleanup() {
  if [ -n "${PF_PID:-}" ]; then
    kill "$PF_PID" 2>/dev/null || true
    wait "$PF_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

check() {
  local desc="$1" expected="$2" method="${3:-GET}" url="$4"
  local output code
  output=$(curl -sf -w "\n%{http_code}" -X "$method" "$url" 2>/dev/null || echo -e "\n000")
  code=$(echo "$output" | tail -1)
  local body
  body=$(echo "$output" | sed '$d')

  if echo "$body" | grep -q "$expected"; then
    echo "  PASS: $desc (HTTP $code)"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: $desc (HTTP $code)"
    echo "    body: $body"
    FAIL=$((FAIL + 1))
  fi
}

echo "=== Observability Verification ==="

echo "--- Metrics ---"
# gateway-api exposes 6060 on host
check "gateway-api /metrics" "go_" GET "http://localhost:6060/metrics"
check "Prometheus healthy"     "Prometheus Server is Healthy" GET "http://localhost:9090/-/healthy"

echo "--- Prometheus Targets ---"
for svc in gateway-api user-rpc product-rpc inventory-rpc order-rpc; do
  result=$(curl -sf "http://localhost:9090/api/v1/query?query=up%7Binstance%3D%22${svc}:6060%22%7D" 2>/dev/null || echo "")
  if echo "$result" | grep -q '"value":\[.*,"1"\]'; then
    echo "  PASS: $svc:6060 target UP"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: $svc:6060 target not UP"
    FAIL=$((FAIL + 1))
  fi
done

echo "--- Jaeger ---"
check "Jaeger UI accessible" "jaeger" GET "http://localhost:16686"

echo "--- Trace ---"
# Use unique values for rerunnable test
SUFFIX=$(date +%s)
# Phase 9: use Kratos auth to create an authenticated order trace.
AUTH_USER="trace-${SUFFIX}"
AUTH_RESP=$(curl -sf -X POST "$BASE/api/v1/auth/register" -H 'Content-Type: application/json' \
  -d "{\"username\":\"${AUTH_USER}\",\"password\":\"123456\"}")
SESSION_TOKEN=$(echo "$AUTH_RESP" | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
PRODUCT=$(curl -sf -X POST "$BASE/api/v1/products" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -d '{"name":"ObsKB","price_cents":100}' | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
curl -sf -X PUT "$BASE/api/v1/inventories/$PRODUCT" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -d '{"stock":5}' >/dev/null
curl -sf -X POST "$BASE/api/v1/orders" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $SESSION_TOKEN" \
  -d "{\"product_id\":$PRODUCT,\"quantity\":1}" >/dev/null
sleep 3

# Check Jaeger has traces for gateway-api
TRACES=$(curl -sf "http://localhost:16686/api/traces?service=gateway-api&limit=1" 2>/dev/null | grep -c '"traceID"' || echo "0")
TRACES=$(echo "$TRACES" | tr -d '[:space:]')
if [ "${TRACES:-0}" -gt 0 ]; then
  echo "  PASS: traces found in Jaeger ($TRACES)"
  PASS=$((PASS + 1))
else
  echo "  FAIL: no traces found"
  FAIL=$((FAIL + 1))
fi

echo "--- Loki ---"
LOKI_READY=$(curl -sf "http://localhost:3100/ready" 2>/dev/null | grep -ci "ready" || echo "0")
LOKI_READY=$(echo "$LOKI_READY" | tr -d '[:space:]')
if [ "${LOKI_READY:-0}" -gt 0 ]; then
  echo "  PASS: Loki ready"
  PASS=$((PASS + 1))
else
  echo "  FAIL: Loki not ready"
  FAIL=$((FAIL + 1))
fi

echo "--- X-Trace-Id ---"
HEALTHZ_HEADERS=$(curl -sD - -o /dev/null "$BASE/healthz" 2>/dev/null || echo "")
if echo "$HEALTHZ_HEADERS" | grep -qi "X-Trace-Id"; then
  echo "  PASS: X-Trace-Id header present"
  PASS=$((PASS + 1))
else
  echo "  FAIL: X-Trace-Id header missing"
  FAIL=$((FAIL + 1))
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
