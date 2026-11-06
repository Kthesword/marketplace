CREATE TABLE stores (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    seller_id     BIGINT      NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    name          TEXT        NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    logo_url      TEXT        NOT NULL DEFAULT '',
    contact_email TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    seller_id   BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    price_cents BIGINT      NOT NULL,
    stock       INT         NOT NULL,
    image_url   TEXT        NOT NULL DEFAULT '',
    active      BOOLEAN     NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT products_price_nonneg CHECK (price_cents >= 0),
    CONSTRAINT products_stock_nonneg CHECK (stock >= 0)
);

CREATE INDEX idx_products_seller_id ON products(seller_id);
CREATE INDEX idx_products_active_created ON products(active, created_at DESC);

