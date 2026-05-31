#!/usr/bin/env bash
set -euo pipefail

BASE="${BASE_URL:-http://localhost:8080}"
PASS=0
FAIL=0

check() {
  local desc="$1" expected="$2"
  local method="${3:-GET}"
  local url="${4:-}"
  local data="${5:-}"

  local curl_args=(-s -w "\n%{http_code}" -X "$method")
  if [ -n "$data" ]; then
    curl_args+=(-H 'Content-Type: application/json' -d "$data")
  fi
  curl_args+=("$BASE$url")

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

echo "=== E2E Compose Verification ==="
echo ""

echo "--- Health ---"
check "healthz ok" '"status":"ok"' GET "/healthz"

echo "--- User ---"
check "register" '"id":1' POST "/api/v1/register" '{"username":"alice","password":"123456"}'
check "login"    '"id":1' POST "/api/v1/login"    '{"username":"alice","password":"123456"}'

echo "--- Product ---"
check "create product" '"status":"active"' POST "/api/v1/products" \
  '{"name":"Keyboard","description":"Mechanical keyboard","price_cents":19900}'
check "list products" '"Keyboard"' GET "/api/v1/products"
check "get product"   '"price_cents":19900' GET "/api/v1/products/1"

echo "--- Inventory ---"
check "set stock" '"stock":10' PUT "/api/v1/inventories/1" '{"stock":10}'
check "get stock" '"stock":10' GET "/api/v1/inventories/1"

echo "--- Order ---"
check "create order" '"total_price_cents":39800' POST "/api/v1/orders" \
  '{"user_id":1,"product_id":1,"quantity":2}'
check "get order" '"status":"created"' GET "/api/v1/orders/1"

echo "--- Post-order Inventory ---"
check "stock deducted to 8" '"stock":8' GET "/api/v1/inventories/1"

echo "--- User Orders ---"
check "user orders" '"total_price_cents":39800' GET "/api/v1/users/1/orders"

echo "--- Negative: Stock Insufficient ---"
neg_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/v1/orders" \
  -H 'Content-Type: application/json' \
  -d '{"user_id":1,"product_id":1,"quantity":999}')
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
