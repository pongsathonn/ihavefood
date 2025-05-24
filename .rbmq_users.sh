#!/bin/bash

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

panda "$RBMQ_ORDER_USER" "$RBMQ_ORDER_PASS"
panda "$RBMQ_RESTAURANT_USER" "$RBMQ_RESTAURANT_PASS"
panda "$RBMQ_PROFILE_USER" "$RBMQ_PROFILE_PASS"
panda "$RBMQ_DELIVERY_USER" "$RBMQ_DELIVERY_PASS"
panda "$RBMQ_COUPON_USER" "$RBMQ_COUPON_PASS"

wait $RABBITMQ_PID
