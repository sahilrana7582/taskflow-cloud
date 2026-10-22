-- A workspace is the tenant boundary: every project and job belongs to exactly one.
CREATE TABLE workspaces (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT workspaces_name_length CHECK (char_length(name) BETWEEN 1 AND 100),
    CONSTRAINT workspaces_slug_format CHECK (
        slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND char_length(slug) BETWEEN 3 AND 50
    )
);

CREATE UNIQUE INDEX workspaces_slug_unique ON workspaces (slug) WHERE deleted_at IS NULL;
CREATE INDEX workspaces_created_by_idx ON workspaces (created_by);

CREATE TRIGGER workspaces_set_updated_at
    BEFORE UPDATE ON workspaces
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE workspace_members (
    workspace_id UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role         TEXT NOT NULL DEFAULT 'member',
    invited_by   UUID REFERENCES users (id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (workspace_id, user_id),
    CONSTRAINT workspace_members_role_valid CHECK (role IN ('owner', 'admin', 'member', 'viewer'))
);

-- "Which workspaces am I in?" (the PK already covers lookups by workspace).
CREATE INDEX workspace_members_user_id_idx ON workspace_members (user_id);

CREATE TRIGGER workspace_members_set_updated_at
    BEFORE UPDATE ON workspace_members
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
