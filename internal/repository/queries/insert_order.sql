insert into wb_demo."order"
    (
     order_uid,
     track_number,
     entry,
     locale,
     internal_signature,
     customer_id,
     delivery_service,
     shardkey,
     sm_id,
     date_created,
     oof_shard
    )
values
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)