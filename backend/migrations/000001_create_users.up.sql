CREATE TABLE users (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,          -- bcrypt hash, never the raw password
    name          TEXT        NOT NULL,
    role          TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Safety nets: the DB rejects bad data even if application code has a bug.
    CONSTRAINT users_role_check  CHECK (role IN ('seller', 'buyer')),
    CONSTRAINT users_email_lower CHECK (email = lower(email))
);

