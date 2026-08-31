package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

var _ ports.Store = (*Store)(nil)

func (s *Store) WithIdempotency(key, fingerprint string, fn func() (any, error)) (any, error) {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	result, err := (&transaction{store: s, tx: tx}).WithIdempotency(key, fingerprint, fn)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) NextID(prefix string) string { return nextSequenceValue(context.Background(), s.db, prefix) }
func (s *Store) NextReference(prefix string) string { return nextSequenceValue(context.Background(), s.db, prefix) }

func (s *Store) SavePaymentIntent(intent domain.PaymentIntent) domain.PaymentIntent {
	_, _ = upsertPaymentIntent(context.Background(), s.db, intent)
	return intent
}

func (s *Store) GetPaymentIntent(id string) (domain.PaymentIntent, error) {
	return getPaymentIntent(context.Background(), s.db, id)
}

func (s *Store) SavePaymentAttempt(attempt domain.PaymentAttempt) domain.PaymentAttempt {
	_, _ = upsertPaymentAttempt(context.Background(), s.db, attempt)
	return attempt
}

func (s *Store) GetPaymentAttempt(id string) (domain.PaymentAttempt, error) {
	return getPaymentAttempt(context.Background(), s.db, id)
}

func (s *Store) SaveCharge(charge domain.Charge) domain.Charge {
	_, _ = upsertCharge(context.Background(), s.db, charge)
	return charge
}

func (s *Store) GetCharge(id string) (domain.Charge, error) { return getCharge(context.Background(), s.db, id) }

func (s *Store) SaveRefund(refund domain.Refund) domain.Refund {
	_, _ = upsertRefund(context.Background(), s.db, refund)
	return refund
}

func (s *Store) GetRefund(id string) (domain.Refund, error) { return getRefund(context.Background(), s.db, id) }

type transaction struct {
	store     *Store
	tx        *sql.Tx
	publisher ports.EventPublisher
	events    []outboxRecord
}

type outboxRecord struct {
	ID    int64
	Event domain.Event
}

var _ ports.Transaction = (*transaction)(nil)

func (tx *transaction) WithIdempotency(key, fingerprint string, fn func() (any, error)) (any, error) {
	var responseType string
	var responsePayload []byte
	var storedFingerprint string
	err := queryRow(tx.tx, `SELECT fingerprint, response_type, response_payload FROM idempotency_keys WHERE key = $1`, key).Scan(&storedFingerprint, &responseType, &responsePayload)
	if err == nil {
		if responseType == "" {
			return nil, domain.NewError(500, "idempotency_corrupt", "stored idempotency record is missing response type")
		}
		if err := compareFingerprints(storedFingerprint, fingerprint); err != nil {
			return nil, err
		}
		return decodeIdempotencyValue(responseType, responsePayload)
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	value, err := fn()
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	responseType = reflect.TypeOf(value).String()
	if responseType == "" {
		return nil, domain.NewError(500, "idempotency_corrupt", "idempotency response type is empty")
	}
	if _, err := exec(tx.tx, `INSERT INTO idempotency_keys(key, fingerprint, response_type, response_payload, created_at) VALUES ($1, $2, $3, $4, now())`, key, fingerprint, responseType, payload); err != nil {
		return nil, err
	}
	return value, nil
}

func (tx *transaction) NextID(prefix string) string { return nextSequenceValue(context.Background(), tx.tx, prefix) }
func (tx *transaction) NextReference(prefix string) string { return nextSequenceValue(context.Background(), tx.tx, prefix) }

func (tx *transaction) SavePaymentIntent(intent domain.PaymentIntent) domain.PaymentIntent {
	_, _ = upsertPaymentIntent(context.Background(), tx.tx, intent)
	return intent
}
func (tx *transaction) GetPaymentIntent(id string) (domain.PaymentIntent, error) {
	return getPaymentIntent(context.Background(), tx.tx, id)
}
func (tx *transaction) SavePaymentAttempt(attempt domain.PaymentAttempt) domain.PaymentAttempt {
	_, _ = upsertPaymentAttempt(context.Background(), tx.tx, attempt)
	return attempt
}
func (tx *transaction) GetPaymentAttempt(id string) (domain.PaymentAttempt, error) {
	return getPaymentAttempt(context.Background(), tx.tx, id)
}
func (tx *transaction) SaveCharge(charge domain.Charge) domain.Charge {
	_, _ = upsertCharge(context.Background(), tx.tx, charge)
	return charge
}
func (tx *transaction) GetCharge(id string) (domain.Charge, error) { return getCharge(context.Background(), tx.tx, id) }
func (tx *transaction) SaveRefund(refund domain.Refund) domain.Refund {
	_, _ = upsertRefund(context.Background(), tx.tx, refund)
	return refund
}
func (tx *transaction) GetRefund(id string) (domain.Refund, error) { return getRefund(context.Background(), tx.tx, id) }

func (tx *transaction) Publish(event domain.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	var id int64
	if err := queryRow(tx.tx, `INSERT INTO outbox_events(event_name, payload, created_at) VALUES ($1, $2, now()) RETURNING id`, event.EventName(), payload).Scan(&id); err != nil {
		return err
	}
	tx.events = append(tx.events, outboxRecord{ID: id, Event: event})
	return nil
}

func (tx *transaction) commit() error {
	if err := tx.tx.Commit(); err != nil {
		return err
	}
	for _, record := range tx.events {
		if tx.publisher != nil {
			if err := tx.publisher.Publish(record.Event); err != nil {
				return err
			}
		}
		_, _ = exec(tx.store.db, `UPDATE outbox_events SET published_at = now() WHERE id = $1`, record.ID)
	}
	return nil
}

type execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func exec(e execer, query string, args ...any) (sql.Result, error) {
	return e.ExecContext(context.Background(), query, args...)
}

func queryRow(e execer, query string, args ...any) *sql.Row {
	return e.QueryRowContext(context.Background(), query, args...)
}

func nextSequenceValue(ctx context.Context, execer execer, prefix string) string {
	var n int64
	_ = queryRow(execer, `SELECT nextval('payment_sandbox_seq')`).Scan(&n)
	return fmt.Sprintf("%s_%06d", prefix, n)
}

func compareFingerprints(existing, incoming string) error {
	if existing != incoming {
		return domain.NewError(409, "idempotency_conflict", "idempotency key was already used with a different request")
	}
	return nil
}

func decodeIdempotencyValue(responseType string, payload []byte) (any, error) {
	switch responseType {
	case "domain.PaymentIntent":
		var v domain.PaymentIntent
		if err := json.Unmarshal(payload, &v); err != nil { return nil, err }
		return v, nil
	case "domain.ConfirmPaymentIntentResponse":
		var v domain.ConfirmPaymentIntentResponse
		if err := json.Unmarshal(payload, &v); err != nil { return nil, err }
		return v, nil
	case "domain.CapturePaymentIntentResponse":
		var v domain.CapturePaymentIntentResponse
		if err := json.Unmarshal(payload, &v); err != nil { return nil, err }
		return v, nil
	case "domain.RefundResponse":
		var v domain.RefundResponse
		if err := json.Unmarshal(payload, &v); err != nil { return nil, err }
		return v, nil
	default:
		return nil, domain.NewError(500, "idempotency_corrupt", "unsupported idempotency response type")
	}
}

func getPaymentIntent(ctx context.Context, q queryer, id string) (domain.PaymentIntent, error) {
	var intent domain.PaymentIntent
	var merchantID, customerID, scenario, idempotencyKey, latestAttemptID, chargeID sql.NullString
	var amount int64
	var currency string
	var createdAt, updatedAt time.Time
	if err := q.QueryRowContext(ctx, `SELECT id, merchant_id, customer_id, amount, currency, capture_method, status, scenario, idempotency_key, latest_attempt_id, charge_id, created_at, updated_at FROM payment_intents WHERE id = $1`, id).Scan(&intent.ID, &merchantID, &customerID, &amount, &currency, &intent.CaptureMethod, &intent.Status, &scenario, &idempotencyKey, &latestAttemptID, &chargeID, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows { return domain.PaymentIntent{}, domain.NewError(404, "payment_intent_not_found", "payment intent not found") }
		return domain.PaymentIntent{}, err
	}
	intent.MerchantID = merchantID.String
	intent.CustomerID = customerID.String
	intent.Amount = domain.Amount(amount)
	intent.Currency = domain.Currency(currency)
	intent.Scenario = scenario.String
	intent.IdempotencyKey = idempotencyKey.String
	intent.LatestAttemptID = latestAttemptID.String
	intent.ChargeID = chargeID.String
	intent.CreatedAt = createdAt
	intent.UpdatedAt = updatedAt
	return intent, nil
}

func upsertPaymentIntent(ctx context.Context, e execer, intent domain.PaymentIntent) (domain.PaymentIntent, error) {
	_, err := exec(e, `INSERT INTO payment_intents(id, merchant_id, customer_id, amount, currency, capture_method, status, scenario, idempotency_key, latest_attempt_id, charge_id, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
ON CONFLICT (id) DO UPDATE SET merchant_id = EXCLUDED.merchant_id, customer_id = EXCLUDED.customer_id, amount = EXCLUDED.amount, currency = EXCLUDED.currency, capture_method = EXCLUDED.capture_method, status = EXCLUDED.status, scenario = EXCLUDED.scenario, idempotency_key = EXCLUDED.idempotency_key, latest_attempt_id = EXCLUDED.latest_attempt_id, charge_id = EXCLUDED.charge_id, created_at = EXCLUDED.created_at, updated_at = EXCLUDED.updated_at`,
		intent.ID, nullString(intent.MerchantID), nullString(intent.CustomerID), int64(intent.Amount), intent.Currency.String(), intent.CaptureMethod, string(intent.Status), nullString(intent.Scenario), nullString(intent.IdempotencyKey), nullString(intent.LatestAttemptID), nullString(intent.ChargeID), intent.CreatedAt, intent.UpdatedAt)
	return intent, err
}

func getPaymentAttempt(ctx context.Context, q queryer, id string) (domain.PaymentAttempt, error) {
	var attempt domain.PaymentAttempt
	var paymentMethodToken, declineCode, processorReference sql.NullString
	var requestedAt, respondedAt time.Time
	if err := q.QueryRowContext(ctx, `SELECT id, payment_intent_id, payment_method_token, status, decline_code, processor_reference, requested_at, responded_at FROM payment_attempts WHERE id = $1`, id).Scan(&attempt.ID, &attempt.PaymentIntentID, &paymentMethodToken, &attempt.Status, &declineCode, &processorReference, &requestedAt, &respondedAt); err != nil {
		if err == sql.ErrNoRows { return domain.PaymentAttempt{}, domain.NewError(404, "payment_attempt_not_found", "payment attempt not found") }
		return domain.PaymentAttempt{}, err
	}
	attempt.PaymentMethodToken = paymentMethodToken.String
	attempt.DeclineCode = declineCode.String
	attempt.ProcessorReference = processorReference.String
	attempt.RequestedAt = requestedAt
	attempt.RespondedAt = respondedAt
	return attempt, nil
}

func upsertPaymentAttempt(ctx context.Context, e execer, attempt domain.PaymentAttempt) (domain.PaymentAttempt, error) {
	_, err := exec(e, `INSERT INTO payment_attempts(id, payment_intent_id, payment_method_token, status, decline_code, processor_reference, requested_at, responded_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
ON CONFLICT (id) DO UPDATE SET payment_intent_id = EXCLUDED.payment_intent_id, payment_method_token = EXCLUDED.payment_method_token, status = EXCLUDED.status, decline_code = EXCLUDED.decline_code, processor_reference = EXCLUDED.processor_reference, requested_at = EXCLUDED.requested_at, responded_at = EXCLUDED.responded_at`,
		attempt.ID, attempt.PaymentIntentID, nullString(attempt.PaymentMethodToken), string(attempt.Status), nullString(attempt.DeclineCode), nullString(attempt.ProcessorReference), attempt.RequestedAt, attempt.RespondedAt)
	return attempt, err
}

func getCharge(ctx context.Context, q queryer, id string) (domain.Charge, error) {
	var charge domain.Charge
	var paymentAttemptID sql.NullString
	var createdAt, updatedAt time.Time
	var amount, capturedAmount, refundedAmount int64
	if err := q.QueryRowContext(ctx, `SELECT id, payment_intent_id, payment_attempt_id, amount, captured_amount, refunded_amount, status, created_at, updated_at FROM charges WHERE id = $1`, id).Scan(&charge.ID, &charge.PaymentIntentID, &paymentAttemptID, &amount, &capturedAmount, &refundedAmount, &charge.Status, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows { return domain.Charge{}, domain.NewError(404, "charge_not_found", "charge not found") }
		return domain.Charge{}, err
	}
	charge.PaymentAttemptID = paymentAttemptID.String
	charge.Amount = domain.Amount(amount)
	charge.CapturedAmount = domain.Amount(capturedAmount)
	charge.RefundedAmount = domain.Amount(refundedAmount)
	charge.CreatedAt = createdAt
	charge.UpdatedAt = updatedAt
	return charge, nil
}

func upsertCharge(ctx context.Context, e execer, charge domain.Charge) (domain.Charge, error) {
	_, err := exec(e, `INSERT INTO charges(id, payment_intent_id, payment_attempt_id, amount, captured_amount, refunded_amount, status, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (id) DO UPDATE SET payment_intent_id = EXCLUDED.payment_intent_id, payment_attempt_id = EXCLUDED.payment_attempt_id, amount = EXCLUDED.amount, captured_amount = EXCLUDED.captured_amount, refunded_amount = EXCLUDED.refunded_amount, status = EXCLUDED.status, created_at = EXCLUDED.created_at, updated_at = EXCLUDED.updated_at`,
		charge.ID, charge.PaymentIntentID, nullString(charge.PaymentAttemptID), int64(charge.Amount), int64(charge.CapturedAmount), int64(charge.RefundedAmount), string(charge.Status), charge.CreatedAt, charge.UpdatedAt)
	return charge, err
}

func getRefund(ctx context.Context, q queryer, id string) (domain.Refund, error) {
	var refund domain.Refund
	var createdAt, updatedAt time.Time
	var amount int64
	if err := q.QueryRowContext(ctx, `SELECT id, charge_id, payment_intent_id, amount, status, created_at, updated_at FROM refunds WHERE id = $1`, id).Scan(&refund.ID, &refund.ChargeID, &refund.PaymentIntentID, &amount, &refund.Status, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows { return domain.Refund{}, domain.NewError(404, "refund_not_found", "refund not found") }
		return domain.Refund{}, err
	}
	refund.Amount = domain.Amount(amount)
	refund.CreatedAt = createdAt
	refund.UpdatedAt = updatedAt
	return refund, nil
}

func upsertRefund(ctx context.Context, e execer, refund domain.Refund) (domain.Refund, error) {
	_, err := exec(e, `INSERT INTO refunds(id, charge_id, payment_intent_id, amount, status, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (id) DO UPDATE SET charge_id = EXCLUDED.charge_id, payment_intent_id = EXCLUDED.payment_intent_id, amount = EXCLUDED.amount, status = EXCLUDED.status, created_at = EXCLUDED.created_at, updated_at = EXCLUDED.updated_at`,
		refund.ID, refund.ChargeID, refund.PaymentIntentID, int64(refund.Amount), string(refund.Status), refund.CreatedAt, refund.UpdatedAt)
	return refund, err
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func nullString(v string) sql.NullString {
	if v == "" { return sql.NullString{} }
	return sql.NullString{String: v, Valid: true}
}
