-- +goose Up
-- Who did what, when. No foreign keys on purpose: audit rows must outlive
-- the users, workspaces and jobs they describe.
CREATE TABLE audit_logs (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id UUID,                          -- NULL for account-level events (login, password change)
    actor_id     UUID,                          -- NULL for system actions
    action       TEXT NOT NULL,                 -- e.g. "job.created", "job.status_changed"
    entity_type  TEXT NOT NULL,                 -- e.g. "job", "project", "user"
    entity_id    UUID NOT NULL,
    changes      JSONB NOT NULL DEFAULT '{}',   -- e.g. {"status": {"from": "todo", "to": "done"}}
    request_id   TEXT,                          -- ties the entry to the request's logs
    ip_address   INET,
    user_agent   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT audit_logs_action_format CHECK (action ~ '^[a-z_]+\.[a-z_]+$')
);

-- History of a single entity ("show this job's activity").
CREATE INDEX audit_logs_entity_idx ON audit_logs (entity_type, entity_id, created_at DESC);
-- Workspace activity feed and admin audit export.
CREATE INDEX audit_logs_workspace_created_idx ON audit_logs (workspace_id, created_at DESC);
CREATE INDEX audit_logs_actor_created_idx ON audit_logs (actor_id, created_at DESC);

-- Append-only: block UPDATE and DELETE at the database level.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION prevent_audit_log_changes() RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_logs is append-only';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER audit_logs_append_only
    BEFORE UPDATE OR DELETE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_log_changes();

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
DROP FUNCTION IF EXISTS prevent_audit_log_changes();
