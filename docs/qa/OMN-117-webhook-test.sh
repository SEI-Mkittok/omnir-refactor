#!/bin/bash
# QA Test Script for OMN-117: Inbound Email Webhook
# Usage: ./docs/qa/OMN-117-webhook-test.sh [base_url]
#
# Prerequisites:
# - API server running (make up-d)
# - Database seeded with test data (make seed)
#
# This script tests the inbound email webhook endpoints for both
# Postmark and Mailgun providers.

set -e

BASE_URL="${1:-http://localhost:8080}"
WEBHOOK_BASE="$BASE_URL/webhooks/email"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "==========================================="
echo "OMN-117 Webhook Test Suite"
echo "==========================================="
echo ""

# Test 1: Postmark webhook - New ticket from known contact
echo -e "${YELLOW}Test 1: Postmark - New ticket from existing contact${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$WEBHOOK_BASE/postmark" \
  -H "Content-Type: application/json" \
  -d '{
    "MessageID": "<test-message-1@example.com>",
    "From": "John Doe <john.doe@example.com>",
    "Subject": "Test ticket from email",
    "TextBody": "This is a test email body that should become the first comment.",
    "HtmlBody": "",
    "Headers": []
  }')

HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
if [ "$HTTP_CODE" = "200" ]; then
  echo -e "${GREEN}✓ Postmark webhook accepted (HTTP 200)${NC}"
else
  echo -e "${RED}✗ Postmark webhook failed (HTTP $HTTP_CODE)${NC}"
  exit 1
fi
echo ""

# Test 2: Postmark webhook - New ticket from unknown contact
echo -e "${YELLOW}Test 2: Postmark - New ticket from unknown email${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$WEBHOOK_BASE/postmark" \
  -H "Content-Type: application/json" \
  -d '{
    "MessageID": "<test-message-2@example.com>",
    "From": "unknown@example.com",
    "Subject": "Ticket from unknown sender",
    "TextBody": "Email from a sender not in the contacts database.",
    "HtmlBody": "",
    "Headers": []
  }')

HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
if [ "$HTTP_CODE" = "200" ]; then
  echo -e "${GREEN}✓ Postmark webhook accepted (HTTP 200)${NC}"
else
  echo -e "${RED}✗ Postmark webhook failed (HTTP $HTTP_CODE)${NC}"
  exit 1
fi
echo ""

# Test 3: Postmark webhook - Reply to existing ticket
echo -e "${YELLOW}Test 3: Postmark - Reply threading (In-Reply-To)${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$WEBHOOK_BASE/postmark" \
  -H "Content-Type: application/json" \
  -d '{
    "MessageID": "<test-message-3@example.com>",
    "From": "john.doe@example.com",
    "Subject": "Re: Test ticket from email",
    "TextBody": "This is a reply to the original ticket.",
    "HtmlBody": "",
    "Headers": [
      {"Name": "In-Reply-To", "Value": "<test-message-1@example.com>"}
    ]
  }')

HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
if [ "$HTTP_CODE" = "200" ]; then
  echo -e "${GREEN}✓ Reply webhook accepted (HTTP 200)${NC}"
else
  echo -e "${RED}✗ Reply webhook failed (HTTP $HTTP_CODE)${NC}"
  exit 1
fi
echo ""

# Test 4: Mailgun webhook - New ticket
echo -e "${YELLOW}Test 4: Mailgun - New ticket${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$WEBHOOK_BASE/mailgun" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "Message-Id=<mailgun-test-1@example.com>" \
  -d "sender=jane@example.com" \
  -d "subject=Mailgun test ticket" \
  -d "body-plain=This is a test from Mailgun." \
  -d "timestamp=1234567890" \
  -d "token=testtoken" \
  -d "signature=fakesignature")

HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
if [ "$HTTP_CODE" = "200" ]; then
  echo -e "${GREEN}✓ Mailgun webhook accepted (HTTP 200)${NC}"
else
  echo -e "${RED}✗ Mailgun webhook failed (HTTP $HTTP_CODE)${NC}"
  exit 1
fi
echo ""

# Test 5: Empty subject handling
echo -e "${YELLOW}Test 5: Empty subject (should default to '(no subject)')${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$WEBHOOK_BASE/postmark" \
  -H "Content-Type: application/json" \
  -d '{
    "MessageID": "<test-message-4@example.com>",
    "From": "john.doe@example.com",
    "Subject": "",
    "TextBody": "Email with no subject line.",
    "HtmlBody": "",
    "Headers": []
  }')

HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
if [ "$HTTP_CODE" = "200" ]; then
  echo -e "${GREEN}✓ Empty subject handled (HTTP 200)${NC}"
else
  echo -e "${RED}✗ Empty subject failed (HTTP $HTTP_CODE)${NC}"
  exit 1
fi
echo ""

echo "==========================================="
echo "Manual Verification Required:"
echo "==========================================="
echo "1. Check database for created tickets:"
echo "   SELECT id, subject, source, contact_id, email_message_id FROM tickets ORDER BY created_at DESC LIMIT 5;"
echo ""
echo "2. Verify ticket source='email' for all webhook-created tickets"
echo ""
echo "3. Verify contact_id is set for Test 1 (john.doe@example.com)"
echo ""
echo "4. Verify contact_id is NULL for Test 2 (unknown@example.com)"
echo "   ⚠️  KNOWN ISSUE: New contacts are NOT auto-created"
echo ""
echo "5. Verify Test 3 created a comment, not a new ticket:"
echo "   SELECT ticket_id, body FROM ticket_comments WHERE body LIKE '%reply to the original%';"
echo ""
echo "6. Check email_message_id is stored for threading:"
echo "   SELECT email_message_id FROM tickets WHERE subject LIKE 'Test ticket%';"
echo ""
