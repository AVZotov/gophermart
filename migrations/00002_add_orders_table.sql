-- +goose Up
CREATE TABLE orders
(
    order_number TEXT PRIMARY KEY,
    user_id      BIGINT      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status       TEXT        NOT NULL DEFAULT 'NEW' CHECK (status IN ('NEW', 'PROCESSING', 'PROCESSED', 'INVALID')),
    accrual_cents BIGINT,
    uploaded_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_id ON orders (user_id);

-- +goose Down
DROP TABLE orders;
