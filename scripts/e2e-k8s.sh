#!/usr/bin/env bash
set -euo pipefail

K8S_NAMESPACE="${K8S_NAMESPACE:-go-svc}"
BASE_URL="${BASE_URL:-http://localhost:8080}"

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
  local desc="$1" expected="$2" method="${3:-GET}" url="$4" data="${5:-}"
  local curl_args=(-s -w "\n%{http_code}" -X "$method")
  if [ -n "$data" ]; then
    curl_args+=(-H 'Content-Type: application/json' -d "$data")
  fi
  curl_args+=("$BASE_URL$url")

  local output code
  output=$(curl "${curl_args[@]}")
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

# Start port-forward if BASE_URL points to localhost and no PF is running
if echo "$BASE_URL" | grep -q "localhost" && ! lsof -i :8080 -sTCP:LISTEN >/dev/null 2>&1; then
  echo "Starting port-forward..."
  kubectl port-forward svc/gateway-api 8080:8080 -n "$K8S_NAMESPACE" &
  PF_PID=$!

  # Wait for port-forward to be ready
  for i in $(seq 1 15); do
    if curl -sf "$BASE_URL/healthz" >/dev/null 2>&1; then
      echo "port-forward ready"
      break
    fi
    if ! kill -0 "$PF_PID" 2>/dev/null; then
      echo "ERROR: port-forward exited unexpectedly"
      exit 1
    fi
    sleep 1
  done
else
  echo "Using existing endpoint: $BASE_URL"
fi

echo "=== E2E K8s Verification ==="
check "healthz"           '"status":"ok"'              GET  "/healthz"
check "register"          '"id":1'                     POST "/api/v1/register" '{"username":"alice","password":"123456"}'
check "login"             '"id":1'                     POST "/api/v1/login"    '{"username":"alice","password":"123456"}'
check "create product"    '"status":"active"'          POST "/api/v1/products" '{"name":"Keyboard","price_cents":19900}'
check "list products"     '"Keyboard"'                 GET  "/api/v1/products"
check "set stock"         '"stock":10'                 PUT  "/api/v1/inventories/1" '{"stock":10}'
check "create order"      '"total_price_cents":39800'  POST "/api/v1/orders"    '{"user_id":1,"product_id":1,"quantity":2}'
check "get order"         '"status":"created"'         GET  "/api/v1/orders/1"
check "stock deducted"    '"stock":8'                  GET  "/api/v1/inventories/1"
check "user orders"       '"total_price_cents":39800'  GET  "/api/v1/users/1/orders"

# Negative: stock insufficient, response should not be 500
echo "--- Negative: stock insufficient ---"
neg_output=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/orders" \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1,"product_id":1,"quantity":999}')
neg_code=$(echo "$neg_output" | tail -1)
if [ "$neg_code" != "500" ] && [ "$neg_code" != "000" ]; then
  echo "  PASS: stock insufficient returned HTTP $neg_code (non-500)"
  PASS=$((PASS + 1))
else
  echo "  FAIL: stock insufficient returned HTTP $neg_code (expected non-500)"
  FAIL=$((FAIL + 1))
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
