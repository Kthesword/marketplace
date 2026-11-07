CREATE TABLE orders (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    checkout_id  UUID        NOT NULL,              -- groups the per-seller orders of one checkout
    buyer_id     BIGINT      NOT NULL REFERENCES users(id),
    seller_id    BIGINT      NOT NULL REFERENCES users(id),
    status       TEXT        NOT NULL DEFAULT 'Pending',
    total_cents  BIGINT      NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT orders_status_check CHECK (status IN ('Pending', 'Shipped', 'Delivered', 'Cancelled')),
    CONSTRAINT orders_total_nonneg CHECK (total_cents >= 0)
);

CREATE INDEX idx_orders_buyer  ON orders(buyer_id, created_at DESC);
CREATE INDEX idx_orders_seller ON orders(seller_id, created_at DESC);
CREATE INDEX idx_orders_checkout ON orders(checkout_id);

CREATE TABLE order_items (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id         BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    -- Nullable + SET NULL: deleting a product must not destroy order history.
    product_id       BIGINT REFERENCES products(id) ON DELETE SET NULL,
    product_name     TEXT   NOT NULL,   -- snapshot at purchase time
    unit_price_cents BIGINT NOT NULL,   -- snapshot at purchase time
    quantity         INT    NOT NULL,

    CONSTRAINT order_items_qty_pos   CHECK (quantity > 0),
    CONSTRAINT order_items_price_pos CHECK (unit_price_cents >= 0)
);

CREATE INDEX idx_order_items_order ON order_items(order_id);
