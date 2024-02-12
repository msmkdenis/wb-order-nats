select
    o.order_uid,
    o.track_number,
    o.entry,
    json_build_object(
            'name', d.name,
            'phone', d.phone,
            'zip', d.zip,
            'city', d.city,
            'address', d.address,
            'region', d.region,
            'email', d.email)
        as delivery,
    json_build_object(
            'transaction', p.transaction,
            'request_id', p.request_id,
            'currency', p.currency,
            'provider', p.provider,
            'amount', p.amount,
            'payment_dt', extract(epoch from p.payment_dt)::integer,
            'bank', p.bank,
            'delivery_cost', p.delivery_cost,
            'goods_total', p.goods_total,
            'custom_fee', p.custom_fee)
        as payment,
    json_agg(json_build_object(
            'chrt_id', i.chrt_id,
            'track_number', i.track_number,
            'price', i.price,
            'rid', i.rid,
            'name', i.name,
            'sale', i.sale,
            'size', i.size,
            'total_price', i.total_price,
            'nm_id', i.nm_id,
            'brand', i.brand,
            'status', i.status))
        as items,
    o.locale,
    o.internal_signature,
    o.customer_id,
    o.delivery_service,
    o.shardkey,
    o.sm_id,
    o.date_created::text,
        o.oof_shard
from wb_demo."order" o
         left join wb_demo.delivery d on o.order_uid = d.order_uid
         left join wb_demo.item i on o.order_uid = i.order_uid
         left join wb_demo.payment p on o.order_uid = p.order_uid
group by o.order_uid,
         d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
         p.transaction, p.request_id, p.currency, p.provider, p.amount, extract(epoch from p.payment_dt)::integer,
         p.bank, p.delivery_cost, p.goods_total, p.custom_fee
