
Place Order Event
+====================================================================+
|   exchange   | type   |     routing key    |         queue         |
|====================================================================|
| order        | topic  | order.placed.event | payment-created-queue |
| order        | topic  | order.placed.event | user=updated=queue    |
+====================================================================+


Order Status Event
+=================================================================+
| exchange     | type   | routing key        | queue              |
|=================================================================|
| order        | topic  | user.event.created | user=created=queue |
| order        | topic  | user.event.updated | user=updated=queue |
| order        | topic  | user.event.deleted | user=deleted=queue |
| order        | topic  | user.cmd.create    | user=create=queue  |
+=================================================================+



