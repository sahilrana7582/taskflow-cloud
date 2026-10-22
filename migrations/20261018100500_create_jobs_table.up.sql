CREATE TABLE jobs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    project_id   UUID NOT NULL,
    number       INTEGER NOT NULL,
    title        TEXT NOT NULL,
    description  TEXT,
    status       TEXT NOT NULL DEFAULT 'todo',
    priority     TEXT NOT NULL DEFAULT 'medium',
    reporter_id  UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    assignee_id  UUID REFERENCES users (id) ON DELETE SET NULL,
    due_at       TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    -- Optimistic locking: UPDATE ... WHERE id = $1 AND version = $2, then bump it.
    -- Zero rows updated means someone else changed the job first (HTTP 409).
    version      INTEGER NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ,

    CONSTRAINT jobs_project_fk FOREIGN KEY (project_id, workspace_id)
        REFERENCES projects (id, workspace_id) ON DELETE CASCADE,
    CONSTRAINT jobs_project_number_unique UNIQUE (project_id, number),
    CONSTRAINT jobs_number_positive CHECK (number > 0),
    CONSTRAINT jobs_title_length CHECK (char_length(title) BETWEEN 1 AND 200),
    CONSTRAINT jobs_description_length CHECK (char_length(description) <= 10000),
    CONSTRAINT jobs_status_valid CHECK (status IN ('todo', 'in_progress', 'blocked', 'done', 'cancelled')),
    CONSTRAINT jobs_priority_valid CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    CONSTRAINT jobs_completed_at_matches_status CHECK ((status = 'done') = (completed_at IS NOT NULL)),
    CONSTRAINT jobs_version_positive CHECK (version > 0)
);

-- Main list endpoint: a project's live jobs, newest first, keyset-paginated on (created_at, id).
CREATE INDEX jobs_project_created_idx ON jobs (project_id, created_at DESC, id DESC) WHERE deleted_at IS NULL;
-- Workspace-wide filters ("all blocked jobs") and FK cascade lookups.
CREATE INDEX jobs_workspace_status_idx ON jobs (workspace_id, status);
-- "My jobs" and FK lookups when a user is removed.
CREATE INDEX jobs_assignee_status_idx ON jobs (assignee_id, status);
CREATE INDEX jobs_reporter_id_idx ON jobs (reporter_id);

CREATE TRIGGER jobs_set_updated_at
    BEFORE UPDATE ON jobs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
