begin transaction;

create schema if not exists wb_demo;

create table if not exists wb_demo.track
(
    track_number            text unique not null,
    constraint pk_track primary key (track_number)
);

create table if not exists wb_demo.order
(
    order_uid               text unique not null,
    track_number            text not null,
    entry                   text not null,
    locale                  text not null,
    internal_signature      text not null,
    customer_id             text not null,
    delivery_service        text not null,
    shardkey                text not null,
    sm_id                   integer not null,
    date_created            timestamp not null,
    oof_shard               text not null,
    constraint pk_order primary key (order_uid),
    constraint fk_track_track_number foreign key (track_number) references wb_demo.track (track_number)
);

create table if not exists wb_demo.payment
(
    transaction             text unique not null,
    order_uid               text unique not null,
    request_id              text not null,
    currency                char(3) not null,
    provider                text not null,
    amount                  integer not null,
    payment_dt              timestamp not null,
    bank                    text not null,
    delivery_cost           integer not null,
    goods_total             integer not null,
    custom_fee              integer not null,
    constraint pk_payment primary key (transaction),
    constraint fk_order_uid foreign key (order_uid) references wb_demo.order (order_uid)
);

create table if not exists wb_demo.item
(
    chrt_id                 bigint unique not null,
    order_uid               text not null,
    track_number            text not null,
    price                   integer not null,
    rid                     text not null,
    name                    text not null,
    sale                    integer not null,
    size                    integer not null,
    total_price             integer not null,
    nm_id                   bigint not null,
    brand                   text not null,
    status                  integer not null,
    constraint pk_item primary key (chrt_id),
    constraint fk_order_uid foreign key (order_uid) references wb_demo.order (order_uid),
    constraint fk_track_track_number foreign key (track_number) references wb_demo.track (track_number)
);

create table if not exists wb_demo.delivery
(
    id                      uuid default gen_random_uuid(),
    order_uid               text unique not null,
    name                    text not null,
    phone                   text not null,
    zip                     text not null,
    city                    text not null,
    address                 text not null,
    region                  text not null,
    email                   text not null,
    constraint pk_delivery primary key (id),
    constraint fk_order_uid foreign key (order_uid) references wb_demo.order (order_uid)
);

commit transaction;