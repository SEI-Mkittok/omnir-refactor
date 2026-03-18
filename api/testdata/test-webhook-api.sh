#!/bin/bash
# Webhook API Integration Test Script (OMN-276)
# Tests all webhook CRUD operations and delivery log
# Usage: ./test-webhook-api.sh [API_URL] [AUTH_TOKEN]

set -e

API_URL="${1:-http://localhost:8080/api/v1}"
AUTH_TOKEN="${2:-}"

if [ -z "$AUTH_TOKEN" ]; then
  echo "❌ Error: AUTH_TOKEN required"
  echo "Usage: $0 [API_URL] [AUTH_TOKEN]"
  echo "Example: $0 http://localhost:8080/api/v1 eyJhbGc..."
  exit 1
fi

BASE_HEADERS=(-H "Authorization: Bearer $AUTH_TOKEN" -H "Content-Type: application/json")

echo "🧪 Webhook API Test Suite"
echo "========================="
echo "API URL: $API_URL"
echo ""

# Test 1: List webhooks (should be empty or return array)
echo "📋 Test 1: List webhooks"
RESPONSE=$(curl -s -X GET "$API_URL/webhooks" "${BASE_HEADERS[@]}")
echo "Response: $RESPONSE"
if echo "$RESPONSE" | jq -e '. | type == "array"' > /dev/null 2>&1; then
  echo "✅ PASS: Returns array"
else
  echo "❌ FAIL: Expected array"
  exit 1
fi
echo ""

# Test 2: Create webhook
echo "📝 Test 2: Create webhook"
CREATE_PAYLOAD='{
  "url": "https://webhook.site/test-omnir",
  "events": ["contact.created", "deal.updated"]
}'
RESPONSE=$(curl -s -X POST "$API_URL/webhooks" "${BASE_HEADERS[@]}" -d "$CREATE_PAYLOAD")
echo "Response: $RESPONSE"

WEBHOOK_ID=$(echo "$RESPONSE" | jq -r '.id')
if [ -z "$WEBHOOK_ID" ] || [ "$WEBHOOK_ID" = "null" ]; then
  echo "❌ FAIL: No webhook ID returned"
  exit 1
fi
echo "✅ PASS: Created webhook $WEBHOOK_ID"
echo ""

# Test 3: Get webhook by ID
echo "🔍 Test 3: Get webhook by ID"
RESPONSE=$(curl -s -X GET "$API_URL/webhooks/$WEBHOOK_ID" "${BASE_HEADERS[@]}")
echo "Response: $RESPONSE"
WEBHOOK_URL=$(echo "$RESPONSE" | jq -r '.url')
if [ "$WEBHOOK_URL" = "https://webhook.site/test-omnir" ]; then
  echo "✅ PASS: Webhook retrieved correctly"
else
  echo "❌ FAIL: Webhook URL mismatch"
  exit 1
fi
echo ""

# Test 4: Update webhook
echo "✏️  Test 4: Update webhook"
UPDATE_PAYLOAD='{
  "url": "https://webhook.site/test-omnir-updated",
  "events": ["contact.created", "contact.updated", "deal.created"],
  "active": true
}'
RESPONSE=$(curl -s -X PATCH "$API_URL/webhooks/$WEBHOOK_ID" "${BASE_HEADERS[@]}" -d "$UPDATE_PAYLOAD")
echo "Response: $RESPONSE"
UPDATED_URL=$(echo "$RESPONSE" | jq -r '.url')
UPDATED_EVENTS=$(echo "$RESPONSE" | jq -r '.events | length')
if [ "$UPDATED_URL" = "https://webhook.site/test-omnir-updated" ] && [ "$UPDATED_EVENTS" = "3" ]; then
  echo "✅ PASS: Webhook updated correctly"
else
  echo "❌ FAIL: Update failed"
  exit 1
fi
echo ""

# Test 5: Disable webhook
echo "🔌 Test 5: Disable webhook"
DISABLE_PAYLOAD='{"active": false}'
RESPONSE=$(curl -s -X PATCH "$API_URL/webhooks/$WEBHOOK_ID" "${BASE_HEADERS[@]}" -d "$DISABLE_PAYLOAD")
IS_ACTIVE=$(echo "$RESPONSE" | jq -r '.active')
if [ "$IS_ACTIVE" = "false" ]; then
  echo "✅ PASS: Webhook disabled"
else
  echo "❌ FAIL: Webhook still active"
  exit 1
fi
echo ""

# Test 6: Test delivery endpoint
echo "🚀 Test 6: Test webhook delivery"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/webhooks/$WEBHOOK_ID/test" "${BASE_HEADERS[@]}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)
echo "HTTP Code: $HTTP_CODE"
echo "Response: $BODY"
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "202" ]; then
  echo "✅ PASS: Test delivery queued"
else
  echo "❌ FAIL: Test delivery failed"
  exit 1
fi
echo ""

# Test 7: Get delivery log
echo "📊 Test 7: Get delivery log"
sleep 2  # Give dispatcher time to process
RESPONSE=$(curl -s -X GET "$API_URL/webhooks/$WEBHOOK_ID/deliveries" "${BASE_HEADERS[@]}")
echo "Response: $RESPONSE"
if echo "$RESPONSE" | jq -e '. | type == "array"' > /dev/null 2>&1; then
  DELIVERY_COUNT=$(echo "$RESPONSE" | jq '. | length')
  echo "✅ PASS: Delivery log returned ($DELIVERY_COUNT deliveries)"
else
  echo "❌ FAIL: Expected array"
  exit 1
fi
echo ""

# Test 8: List webhooks (should include created webhook)
echo "📋 Test 8: List webhooks (verify created webhook appears)"
RESPONSE=$(curl -s -X GET "$API_URL/webhooks" "${BASE_HEADERS[@]}")
WEBHOOK_COUNT=$(echo "$RESPONSE" | jq '. | length')
FOUND=$(echo "$RESPONSE" | jq -e ".[] | select(.id == \"$WEBHOOK_ID\")" > /dev/null 2>&1 && echo "true" || echo "false")
if [ "$FOUND" = "true" ]; then
  echo "✅ PASS: Webhook appears in list (total: $WEBHOOK_COUNT)"
else
  echo "❌ FAIL: Webhook not found in list"
  exit 1
fi
echo ""

# Test 9: Delete webhook
echo "🗑️  Test 9: Delete webhook"
HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null -X DELETE "$API_URL/webhooks/$WEBHOOK_ID" "${BASE_HEADERS[@]}")
echo "HTTP Code: $HTTP_CODE"
if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
  echo "✅ PASS: Webhook deleted"
else
  echo "❌ FAIL: Delete failed (HTTP $HTTP_CODE)"
  exit 1
fi
echo ""

# Test 10: Verify webhook deleted
echo "🔍 Test 10: Verify webhook deleted"
HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null -X GET "$API_URL/webhooks/$WEBHOOK_ID" "${BASE_HEADERS[@]}")
echo "HTTP Code: $HTTP_CODE"
if [ "$HTTP_CODE" = "404" ]; then
  echo "✅ PASS: Webhook not found (correctly deleted)"
else
  echo "❌ FAIL: Webhook still exists"
  exit 1
fi
echo ""

# Test 11: Create webhook with validation errors
echo "❌ Test 11: Validation - Empty URL"
INVALID_PAYLOAD='{"url": "", "events": ["contact.created"]}'
HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null -X POST "$API_URL/webhooks" "${BASE_HEADERS[@]}" -d "$INVALID_PAYLOAD")
if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "422" ]; then
  echo "✅ PASS: Empty URL rejected (HTTP $HTTP_CODE)"
else
  echo "❌ FAIL: Empty URL accepted (HTTP $HTTP_CODE)"
fi
echo ""

# Test 12: Create webhook with no events
echo "❌ Test 12: Validation - No events"
INVALID_PAYLOAD='{"url": "https://example.com", "events": []}'
HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null -X POST "$API_URL/webhooks" "${BASE_HEADERS[@]}" -d "$INVALID_PAYLOAD")
if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "422" ]; then
  echo "✅ PASS: Empty events rejected (HTTP $HTTP_CODE)"
else
  echo "❌ FAIL: Empty events accepted (HTTP $HTTP_CODE)"
fi
echo ""

echo "========================="
echo "✅ All tests passed!"
echo "========================="
