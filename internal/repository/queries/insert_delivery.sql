insert into wb_demo.delivery
    (
     name,
     phone,
     zip,
     city,
     address,
     region,
     email,
     order_uid)
values
    ($1, $2, $3, $4, $5, $6, $7, $8)