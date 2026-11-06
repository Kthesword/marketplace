CREATE TABLE cart_items (
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id BIGINT      NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity   INT         NOT NULL,
    added_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (user_id, product_id),
    CONSTRAINT cart_items_quantity_pos CHECK (quantity > 0)
);

