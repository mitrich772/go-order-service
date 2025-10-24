
CREATE TABLE orders (
    order_uid VARCHAR(36) PRIMARY KEY,
    customer_id VARCHAR(50) NOT NULL,
    locale CHAR(2) NOT NULL,
    delivery_service VARCHAR(50) NOT NULL,
    shard_key VARCHAR(10) NOT NULL,
    sm_id SMALLINT CHECK (sm_id > 0),
    date_created TIMESTAMPTZ NOT NULL,
    oof_shard VARCHAR(10) NOT NULL,
    track_number VARCHAR(50) NOT NULL,
    entry VARCHAR(20) NOT NULL,
    internal_signature VARCHAR(255)
);

CREATE TABLE deliveries (
    delivery_id BIGSERIAL PRIMARY KEY,
    order_uid VARCHAR(36) UNIQUE REFERENCES orders(order_uid),
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    zip VARCHAR(20) NOT NULL,
    city VARCHAR(50) NOT NULL,
    address VARCHAR(200) NOT NULL,
    region VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL
);

CREATE TABLE payments (
    payment_id BIGSERIAL PRIMARY KEY,
    order_uid VARCHAR(36) UNIQUE REFERENCES orders(order_uid),
    transaction VARCHAR(36) NOT NULL,
    request_id VARCHAR(50),
    currency CHAR(3) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    amount NUMERIC(12,2) CHECK (amount >= 0),
    payment_dt BIGINT,
    bank VARCHAR(50),
    delivery_cost NUMERIC(12,2) CHECK (delivery_cost >= 0),
    goods_total NUMERIC(12,2) CHECK (goods_total >= 0),
    custom_fee NUMERIC(12,2) CHECK (custom_fee >= 0)
);

CREATE TABLE items (
    item_id BIGSERIAL PRIMARY KEY,
    order_uid VARCHAR(36) REFERENCES orders(order_uid),
    chrt_id BIGINT CHECK (chrt_id > 0),
    track_number VARCHAR(50) NOT NULL,
    price NUMERIC(12,2) CHECK (price >= 0),
    rid VARCHAR(36),
    name VARCHAR(200) NOT NULL,
    sale NUMERIC(5,2) CHECK (sale >= 0),
    size VARCHAR(10),
    total_price NUMERIC(12,2) CHECK (total_price >= 0),
    nm_id BIGINT CHECK (nm_id > 0),
    brand VARCHAR(100),
    status SMALLINT
);