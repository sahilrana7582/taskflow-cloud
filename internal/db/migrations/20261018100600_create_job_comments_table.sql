-- +goose Up
CREATE TABLE job_comments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id     UUID NOT NULL REFERENCES jobs (id) ON DELETE CASCADE,
    author_id  UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    body       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT job_comments_body_length CHECK (char_length(body) BETWEEN 1 AND 10000)
);

-- A job's comment thread, oldest first.
CREATE INDEX job_comments_job_created_idx ON job_comments (job_id, created_at, id);
CREATE INDEX job_comments_author_id_idx ON job_comments (author_id);

CREATE TRIGGER job_comments_set_updated_at
    BEFORE UPDATE ON job_comments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE IF EXISTS job_comments;
