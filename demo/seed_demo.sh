#!/bin/bash
set -ex

URL="https://api-gateway-731964455549.asia-southeast1.run.app"

# ==========================================================
# 1. Create customer & rider
CUSTOMER_RES=$(curl -sf -X POST \
  "$URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "$(jq -c '{email, password, role}' demo_customer.json)")
CUSTOMER_ID=$(echo "$CUSTOMER_RES" | jq -r '.id')
echo "INFO: created customer"

RIDER_RES=$(curl -sf -X POST \
  "$URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "$(jq -c '{email, password, role}' demo_rider.json)")
RIDER_ID=$(echo "$RIDER_RES" | jq -r '.id')
echo "INFO: created rider"

sleep 1
# ==========================================================
# 2. Update address
curl -sf -X POST "$URL/api/customers/$CUSTOMER_ID/address" \
  -H "Content-Type: application/json" \
  -d "$(jq -c --arg id "$CUSTOMER_ID" '{customer_id: $id, address: .address}' demo_customer.json)" > /dev/null

echo "INFO: updated address"

sleep 1
# ==========================================================
# 3. Seed merchants
jq -c '.[]' demo_merchants.json | while read -r merchant; do
  curl -sf -X POST "$URL/api/merchants" \
    -H "Content-Type: application/json" \
    -d "$merchant" > /dev/null
done

echo "INFO: created merchants"

# ==========================================================
# 4. save merchants images to firebase storage
#
# command check 
# $ gsutil ls gs://<URL>.firebasestorage.app/
BUCKET="gs://ihavefood-ee231.firebasestorage.app"

for img in images/*; do
  filename=$(basename "$img")
  mime=$(file --mime-type -b "$img")
  size=$(stat -f%z "$img")
  gsutil -q -h "Content-Type:$mime" \
         -h "x-goog-meta-size:$size" \
         cp "$img" "$BUCKET/images/$filename" >/dev/null
  echo "INFO: uploaded $filename"
done
sleep 1

# ==========================================================
# 5. Create coupons

jq -c '.[]' demo_coupons.json | while read item; do
  curl -sf -X POST "$URL/api/coupons" \
    -H "Content-Type: application/json" \
    -d "$item" > /dev/null
done
echo "INFO: created coupons"

sleep 1

# ===========
# manual command
 # grpcurl \
 #  -proto <SERVICE>.proto \
 #  -d '{"TO":"DO"}' \
 #  -authority <GCP_SERVICE>.run.app \
 #  -H "authorization: Bearer $(gcloud auth print-identity-token)" \
 #  <GCP_SERVICE>.run.app:443 \
 #  ihavefood.Service/YourMethod

# ==========================================================
echo "Seeding script successful"
