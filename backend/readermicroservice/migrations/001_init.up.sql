CREATE TABLE IF NOT EXISTS orders (
  order_uid VARCHAR(100) PRIMARY KEY,
  track_number VARCHAR(100),
  entry VARCHAR(100),
  locale VARCHAR(100),
  internal_signature VARCHAR(100),
  customer_id VARCHAR(100),
  delivery_service VARCHAR(100),
  shardkey VARCHAR(100),
  sm_id INTEGER,
  date_created TIMESTAMP,
  oof_shard VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS delivery (
  id SERIAL PRIMARY KEY,
  order_uid VARCHAR(100) REFERENCES orders(order_uid),
  name VARCHAR(100),
  phone VARCHAR(100),
  zip VARCHAR(100),
  city VARCHAR(100),
  address VARCHAR(255),
  region VARCHAR(100),
  email VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS payments (
  id SERIAL PRIMARY KEY,
  order_uid VARCHAR(100) REFERENCES orders(order_uid),
  transaction VARCHAR(100),
  request_id VARCHAR(100),
  currency VARCHAR(10),
  provider VARCHAR(100),
  amount INTEGER,
  payment_dt BIGINT,
  bank VARCHAR(100),
  delivery_cost INTEGER,
  goods_total INTEGER,
  custom_fee INTEGER
);

CREATE TABLE IF NOT EXISTS items (
  id SERIAL PRIMARY KEY,
  order_uid VARCHAR(100) REFERENCES orders(order_uid),
  chrt_id INTEGER,
  track_number VARCHAR(100),
  price INTEGER,
  rid VARCHAR(100),
  name VARCHAR(255),
  sale INTEGER,
  size VARCHAR(50),
  total_price INTEGER,
  nm_id INTEGER,
  brand VARCHAR(100),
  status INTEGER
);