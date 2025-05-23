#!/bin/sh
set -e
rabbitmq-server &
RABBITMQ_PID=$!
rabbitmqctl wait /var/lib/rabbitmq/mnesia/rabbit@$(hostname).pid --timeout 60

rabbitmqctl add_vhost /
rabbitmqctl add_user "$RBMQ_ORDER_USER" "$RBMQ_ORDER_PASS"
rabbitmqctl add_user "$RBMQ_RESTAURANT_USER" "$RBMQ_RESTAURANT_PASS"
rabbitmqctl add_user "$RBMQ_PROFILE_USER" "$RBMQ_PROFILE_PASS"
rabbitmqctl add_user "$RBMQ_AUTH_USER" "$RBMQ_AUTH_PASS"
rabbitmqctl add_user "$RBMQ_DELIVERY_USER" "$RBMQ_DELIVERY_PASS"
rabbitmqctl add_user "$RBMQ_COUPON_USER" "$RBMQ_COUPON_PASS"

for user in \
  "$RBMQ_ORDER_USER" \
  "$RBMQ_RESTAURANT_USER" \
  "$RBMQ_PROFILE_USER" \
  "$RBMQ_AUTH_USER" \
  "$RBMQ_DELIVERY_USER" \
  "$RBMQ_COUPON_USER"
do
  rabbitmqctl set_permissions -p / "$user" ".*" ".*" ".*"
done

wait $RABBITMQ_PID
