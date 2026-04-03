-- +goose Up
CREATE TABLE login_attempts (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ip_addr      INET NOT NULL,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_login_attempts_ip_time ON login_attempts (ip_addr, attempted_at);

-- +goose Down
DROP TABLE IF EXISTS login_attempts;
