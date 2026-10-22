CREATE TABLE users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email             TEXT NOT NULL,
    password_hash     TEXT NOT NULL,
    full_name         TEXT NOT NULL,
    status            TEXT NOT NULL DEFAULT 'active',
    email_verified_at TIMESTAMPTZ,
    last_login_at     TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,

    -- The app must lowercase emails before writing; the DB rejects anything else.
    CONSTRAINT users_email_format CHECK (email = lower(email) AND position('@' IN email) > 1),
    CONSTRAINT users_full_name_length CHECK (char_length(full_name) BETWEEN 1 AND 100),
    CONSTRAINT users_status_valid CHECK (status IN ('active', 'suspended'))
);

-- Unique among live accounts only, so a soft-deleted account frees its email.
CREATE UNIQUE INDEX users_email_unique ON users (email) WHERE deleted_at IS NULL;

CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
