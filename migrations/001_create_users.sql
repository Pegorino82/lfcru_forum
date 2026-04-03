-- +goose Up
CREATE TABLE users (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username   TEXT NOT NULL,
    email      TEXT NOT NULL,
    pass_hash  BYTEA NOT NULL,
    role       TEXT NOT NULL DEFAULT 'user',
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_users_email    ON users (lower(email));
CREATE UNIQUE INDEX idx_users_username ON users (lower(username));

-- +goose Down
DROP TABLE IF EXISTS users;
