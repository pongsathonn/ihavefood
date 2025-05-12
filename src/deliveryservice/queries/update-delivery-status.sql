UPDATE deliveries
SET 
    status = ?2,
    accept_time = CASE WHEN ?2 = 1 THEN datetime('now') END,
    deliver_time = CASE WHEN ?2 = 2 THEN datetime('now') END
WHERE order_id=?1;

