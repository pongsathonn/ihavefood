#!/bin/bash

set -e

# ensure that rabbitmq server is started before running commands.
rabbitmq-server &
RABBITMQ_PID=$!
rabbitmqctl wait /var/lib/rabbitmq/mnesia/rabbit@$(hostname).pid --timeout 60

rabbitmqctl add_vhost /

panda() {
  if ! rabbitmqctl list_users | grep -q "$1"; then
    rabbitmqctl add_user "$1" "$2"
    rabbitmqctl set_permissions -p / "$1" ".*" ".*" ".*"
  fi
}

panda "$RBMQ_AUTH_USER" "$RBMQ_AUTH_PASS"
panda "$RBMQ_ORDER_USER" "$RBMQ_ORDER_PASS"
panda "$RBMQ_MERCHANT_USER" "$RBMQ_MERCHANT_PASS"
panda "$RBMQ_CUSTOMER_USER" "$RBMQ_CUSTOMER_PASS"
panda "$RBMQ_DELIVERY_USER" "$RBMQ_DELIVERY_PASS"
panda "$RBMQ_COUPON_USER" "$RBMQ_COUPON_PASS"

wait $RABBITMQ_PID
