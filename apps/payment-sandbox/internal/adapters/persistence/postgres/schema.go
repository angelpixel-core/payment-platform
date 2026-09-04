package postgres

import (
	"context"
	"database/sql"
	"strings"
)

const schemaSQL = `
CREATE SEQUENCE IF NOT EXISTS payment_sandbox_seq;

CREATE TABLE IF NOT EXISTS payment_intents (
	id text PRIMARY KEY,
	merchant_id text NULL,
	customer_id text NULL,
	amount bigint NOT NULL,
	currency text NOT NULL,
	capture_method text NOT NULL,
	status text NOT NULL,
	scenario text NULL,
	idempotency_key text NULL,
	latest_attempt_id text NULL,
	charge_id text NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS payment_attempts (
	id text PRIMARY KEY,
	payment_intent_id text NOT NULL,
	payment_method_token text NULL,
	status text NOT NULL,
	decline_code text NULL,
	processor_reference text NULL,
	requested_at timestamptz NOT NULL,
	responded_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS charges (
	id text PRIMARY KEY,
	payment_intent_id text NOT NULL,
	payment_attempt_id text NULL,
	amount bigint NOT NULL,
	captured_amount bigint NOT NULL,
	refunded_amount bigint NOT NULL,
	status text NOT NULL,
	created_at timestamptz NOT NULL,
	captured_at timestamptz NULL,
	updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS refunds (
	id text PRIMARY KEY,
	charge_id text NOT NULL,
	payment_intent_id text NOT NULL,
	amount bigint NOT NULL,
	status text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
	key text PRIMARY KEY,
	fingerprint text NOT NULL,
	response_type text NOT NULL,
	response_payload jsonb NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS outbox_events (
	id bigserial PRIMARY KEY,
	event_name text NOT NULL,
	payload jsonb NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	published_at timestamptz NULL
);
`

func EnsureSchema(ctx context.Context, db *sql.DB) error {
	for _, stmt := range strings.Split(schemaSQL, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}
