package models

import (
	"encoding/json"
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// AuditLog is an append-only record of who did what. Action is
// "<entity>.<verb>", e.g. "job.status_changed".
type AuditLog struct {
	ID          int64           `db:"id" json:"id"`
	WorkspaceID *uuid.UUID      `db:"workspace_id" json:"workspace_id,omitempty"`
	ActorID     *uuid.UUID      `db:"actor_id" json:"actor_id,omitempty"`
	Action      string          `db:"action" json:"action"`
	EntityType  string          `db:"entity_type" json:"entity_type"`
	EntityID    uuid.UUID       `db:"entity_id" json:"entity_id"`
	Changes     json.RawMessage `db:"changes" json:"changes"`
	RequestID   *string         `db:"request_id" json:"request_id,omitempty"`
	IPAddress   *netip.Prefix   `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent   *string         `db:"user_agent" json:"user_agent,omitempty"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
}
