insert into wb_demo.item
    (
     chrt_id,
     track_number,
     price,
     rid,
     name,
     sale,
     size,
     total_price,
     nm_id,
     brand,
     status,
     order_uid
    )
values
    ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)