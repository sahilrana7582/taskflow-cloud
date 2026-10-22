CREATE TABLE projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    key             TEXT NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT,
    -- Per-project counter for human-readable job references (ENG-1, ENG-2, ...).
    -- Incremented with UPDATE ... RETURNING inside the job-insert transaction.
    next_job_number INTEGER NOT NULL DEFAULT 1,
    created_by      UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    archived_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,

    -- Target for the composite FK on jobs, which stops a job in workspace A
    -- from pointing at a project in workspace B.
    CONSTRAINT projects_id_workspace_unique UNIQUE (id, workspace_id),
    CONSTRAINT projects_key_format CHECK (key ~ '^[A-Z][A-Z0-9]{1,9}$'),
    CONSTRAINT projects_name_length CHECK (char_length(name) BETWEEN 1 AND 100),
    CONSTRAINT projects_description_length CHECK (char_length(description) <= 5000),
    CONSTRAINT projects_next_job_number_positive CHECK (next_job_number > 0)
);

CREATE UNIQUE INDEX projects_workspace_key_unique ON projects (workspace_id, key) WHERE deleted_at IS NULL;
CREATE INDEX projects_created_by_idx ON projects (created_by);

CREATE TRIGGER projects_set_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
