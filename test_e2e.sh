#!/bin/bash
BASE_URL="https://valt.turbo.ai.vn"
TEST_EMAIL="testrun-$(date +%s)@valt.dev"
TEST_PASSWORD="TestRun2026!"
REGION_CODE="vn"

TOTAL=0
PASSED=0
FAILED=0

# Register
echo -n "1.  POST /api/v1/auth/register ...................... "
resp=$(curl -s -w "%{http_code}" -X POST "${BASE_URL}/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\",\"region_code\":\"$REGION_CODE\"}")
code=${resp: -3}
TOTAL=$((TOTAL+1))
if [ "$code" = "201" ] || [ "$code" = "200" ]; then PASSED=$((PASSED+1)); echo "PASS"; else FAILED=$((FAILED+1)); echo "FAIL"; fi

# Login
echo -n "2.  POST /api/v1/auth/login ......................... "
resp=$(curl -s -X POST "${BASE_URL}/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
code=$(curl -s -w "%{http_code}" -o /dev/null -X POST "${BASE_URL}/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
TOTAL=$((TOTAL+1))
AUTH_TOKEN=$(echo "$resp" | grep -o '"access_token":"[^"]*' | head -1 | cut -d'"' -f4)
if [ "$code" = "200" ]; then PASSED=$((PASSED+1)); echo "PASS"; else FAILED=$((FAILED+1)); echo "FAIL"; fi

# Get orgs
echo -n "3.  GET /api/v1/orgs ................................ "
code=$(curl -s -w "%{http_code}" -o /dev/null -X GET "${BASE_URL}/api/v1/orgs" \
  -H "Authorization: Bearer $AUTH_TOKEN")
TOTAL=$((TOTAL+1))
if [ "$code" = "200" ]; then PASSED=$((PASSED+1)); echo "PASS"; else FAILED=$((FAILED+1)); echo "FAIL"; fi

# Health check
echo -n "4.  GET /health .................................... "
code=$(curl -s -w "%{http_code}" -o /dev/null -X GET "${BASE_URL}/health")
TOTAL=$((TOTAL+1))
if [ "$code" = "200" ]; then PASSED=$((PASSED+1)); echo "PASS"; else FAILED=$((FAILED+1)); echo "FAIL"; fi

echo ""
echo "Total: $TOTAL | Passed: $PASSED | Failed: $FAILED"
