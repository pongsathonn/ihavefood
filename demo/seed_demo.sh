#!/bin/bash
set -e

# Colors
GREEN='\033[32m'
RED='\033[31m'
NC='\033[0m' # No Color

echo -e "${GREEN}==> (1) Creating customer...${NC}"

REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
  'http://localhost/auth/register' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d "$(jq -c '{email, password, role}' demo_customer.json)")

HTTP_CODE=$(echo "$REGISTER_RESPONSE" | tail -n1)
BODY=$(echo "$REGISTER_RESPONSE" | sed '$d')


if [[ "$HTTP_CODE" == "409" ]]; then
  echo -e "${RED}⚠ Customer already exists (409), skipping creation${NC}"
  CUSTOMER_ID=$(jq -r '.id' <<<"$BODY")
elif [[ "$HTTP_CODE" =~ ^2 ]]; then
  CUSTOMER_ID=$(jq -r '.id' <<<"$BODY")
  echo -e "${GREEN}    ✓ Customer created: $CUSTOMER_ID${NC}"
else
  echo -e "${RED}ERROR: Failed to create customer (HTTP $HTTP_CODE)${NC}"
  echo "Response: $BODY"
  exit 1
fi

echo -e "${GREEN}==> (2) Updating customer address...${NC}"
ADDRESS_BODY=$(jq -c --arg id "$CUSTOMER_ID" '{customer_id: $id, address: .address}' demo_customer.json)
if ! curl -f -o /dev/null -X POST "http://localhost/api/customers/$CUSTOMER_ID/address" \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d "$ADDRESS_BODY"; then
  echo -e "${RED}ERROR: Failed to update address${NC}"
  exit 1
fi
echo -e "${GREEN}    ✓ Address updated${NC}"

echo -e "${GREEN}==> (3) Seeding merchants...${NC}"
DEMO_MERCHANTS="demo_merchants.json"
jq -c '.[]' "$DEMO_MERCHANTS" | while read -r merchant; do
  curl -f -o /dev/null -X POST 'http://localhost/api/merchants' \
    -H 'accept: application/json' \
    -H 'Content-Type: application/json' \
    -d "$merchant" || {
      echo -e "${RED}ERROR: Seeding merchants failed${NC}"
      exit 1
    }
done
echo -e "${GREEN}    ✓ Merchants created${NC}"

# echo -e "${GREEN}✅ All seeding completed!${NC}"
# (4) seeding riders.

