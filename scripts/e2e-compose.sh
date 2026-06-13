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
  local auth_header="${6:-}"

  local curl_args=(-s -w "\n%{http_code}" -X "$method")
  if [ -n "$auth_header" ]; then
    curl_args+=(-H "$auth_header")
  fi
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

echo "--- Auth (Kratos) ---"
# Use a unique username to avoid collisions with legacy users.
AUTH_USER="alice-kratos-$$"
AUTH_PASS="123456"

auth_output=$(curl -s -X POST "$BASE/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"${AUTH_USER}\",\"password\":\"${AUTH_PASS}\"}")
if echo "$auth_output" | grep -q '"session_token"'; then
  echo "  PASS: auth register (200)"
  PASS=$((PASS + 1))
else
  echo "  FAIL: auth register"
  echo "    body: $auth_output"
  FAIL=$((FAIL + 1))
fi

SESSION_TOKEN=$(echo "$auth_output" | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)

# Verify register returns a real local user id.
if echo "$auth_output" | grep -q '"user_id":[1-9]'; then
  echo "  PASS: register returns real user id"
  PASS=$((PASS + 1))
else
  echo "  FAIL: register did not return a real user id"
  echo "    body: $auth_output"
  FAIL=$((FAIL + 1))
fi

login_output=$(curl -s -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"${AUTH_USER}\",\"password\":\"${AUTH_PASS}\"}")
if echo "$login_output" | grep -q '"session_token"'; then
  echo "  PASS: auth login (200)"
  PASS=$((PASS + 1))
  SESSION_TOKEN=$(echo "$login_output" | grep -o '"session_token":"[^"]*"' | cut -d'"' -f4)
else
  echo "  FAIL: auth login"
  FAIL=$((FAIL + 1))
fi

# Verify login returns a real local user id.
if echo "$login_output" | grep -q '"user_id":[1-9]'; then
  echo "  PASS: login returns real user id"
  PASS=$((PASS + 1))
else
  echo "  FAIL: login did not return a real user id"
  echo "    body: $login_output"
  FAIL=$((FAIL + 1))
fi

if [ -n "${SESSION_TOKEN:-}" ]; then
  me_output=$(curl -s -H "Authorization: Bearer $SESSION_TOKEN" "$BASE/api/v1/auth/me")
  if echo "$me_output" | grep -q '"user_id":[1-9]'; then
    echo "  PASS: auth me returns real user id (200)"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: auth me did not return real user id"
    echo "    body: $me_output"
    FAIL=$((FAIL + 1))
  fi
else
  echo "  FAIL: no session token for /me test"
  FAIL=$((FAIL + 1))
fi

AUTH_HEADER="Authorization: Bearer $SESSION_TOKEN"

echo "--- Product ---"
# Unauthenticated product creation should fail
unauth_product_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/v1/products" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Keyboard","description":"Mechanical keyboard","price_cents":19900}')
if [ "$unauth_product_code" = "401" ]; then
  echo "  PASS: create product without token → 401"
  PASS=$((PASS + 1))
else
  echo "  FAIL: create product without token returned HTTP $unauth_product_code (expected 401)"
  FAIL=$((FAIL + 1))
fi

check "create product (auth)" '"status":"active"' POST "/api/v1/products" \
  '{"name":"Keyboard","description":"Mechanical keyboard","price_cents":19900}' "$AUTH_HEADER"
check "list products" '"Keyboard"' GET "/api/v1/products"
check "get product"   '"price_cents":19900' GET "/api/v1/products/1"

echo "--- Inventory ---"
# Unauthenticated stock set should fail
unauth_stock_code=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "$BASE/api/v1/inventories/1" \
  -H 'Content-Type: application/json' \
  -d '{"stock":10}')
if [ "$unauth_stock_code" = "401" ]; then
  echo "  PASS: set stock without token → 401"
  PASS=$((PASS + 1))
else
  echo "  FAIL: set stock without token returned HTTP $unauth_stock_code (expected 401)"
  FAIL=$((FAIL + 1))
fi

check "set stock (auth)" '"stock":10' PUT "/api/v1/inventories/1" '{"stock":10}' "$AUTH_HEADER"
check "get stock" '"stock":10' GET "/api/v1/inventories/1"

echo "--- Order (Auth) ---"
# Unauthenticated order should fail
unauth_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/v1/orders" \
  -H 'Content-Type: application/json' \
  -d '{"product_id":1,"quantity":2}')
if [ "$unauth_code" = "401" ]; then
  echo "  PASS: order without token → 401"
  PASS=$((PASS + 1))
else
  echo "  FAIL: order without token returned HTTP $unauth_code (expected 401)"
  FAIL=$((FAIL + 1))
fi

# Authenticated order with session token
if [ -n "${SESSION_TOKEN:-}" ]; then
  check "create order (auth)" '"total_price_cents":39800' POST "/api/v1/orders" \
    '{"product_id":1,"quantity":2}' "Authorization: Bearer $SESSION_TOKEN"
else
  echo "  FAIL: no session token for authenticated order"
  FAIL=$((FAIL + 1))
fi
check "get order (auth)" '"status":"created"' GET "/api/v1/orders/1" "" "$AUTH_HEADER"

echo "--- Post-order Inventory ---"
check "stock deducted to 8" '"stock":8' GET "/api/v1/inventories/1"

echo "--- Negative: Stock Insufficient ---"
neg_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/v1/orders" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${SESSION_TOKEN:-}" \
  -d '{"product_id":1,"quantity":999}')
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
