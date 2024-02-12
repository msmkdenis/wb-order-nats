insert into wb_demo.payment
    (
     transaction,
     request_id,
     currency,
     provider,
     amount,
     payment_dt,
     bank,
     delivery_cost,
     goods_total,
     custom_fee,
     order_uid
    )
values
    ($1, $2, $3, $4, $5, to_timestamp($6), $7, $8, $9, $10, $11)