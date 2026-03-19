package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type Event struct {
	OrgID       string
	ActorUserID string
	EventType   string
	EntityType  string
	EntityID    string
	Payload     any
}

func WriteTx(ctx context.Context, tx *sql.Tx, event Event) error {
	payload := []byte(`{}`)
	if event.Payload != nil {
		encoded, err := json.Marshal(event.Payload)
		if err != nil {
			return fmt.Errorf("marshal audit payload: %w", err)
		}
		payload = encoded
	}

	const statement = `
INSERT INTO platform.audit_events (
	org_id,
	actor_user_id,
	event_type,
	entity_type,
	entity_id,
	payload
) VALUES ($1, $2, $3, $4, $5, $6::jsonb);`

	_, err := tx.ExecContext(
		ctx,
		statement,
		nullIfEmpty(event.OrgID),
		nullIfEmpty(event.ActorUserID),
		event.EventType,
		event.EntityType,
		event.EntityID,
		string(payload),
	)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}

	return nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
