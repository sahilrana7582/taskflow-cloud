-- +goose Up
-- One row per refresh token. Only the SHA-256 hash is stored, so a DB leak
-- does not leak usable tokens.
CREATE TABLE sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash   BYTEA NOT NULL UNIQUE,
    user_agent   TEXT,
    ip_address   INET,
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked_at   TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT sessions_expires_after_created CHECK (expires_at > created_at)
);

-- "Log out everywhere" and "list my active sessions".
CREATE INDEX sessions_user_id_idx ON sessions (user_id) WHERE revoked_at IS NULL;
-- Cleanup job that deletes expired sessions.
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);

-- +goose Down
DROP TABLE IF EXISTS sessions;
