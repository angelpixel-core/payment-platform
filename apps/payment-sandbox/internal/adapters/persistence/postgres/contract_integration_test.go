//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"payment-sandbox/internal/adapters/messaging/inprocess"
	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type persistenceMetricCall struct {
	backend   string
	resource  string
	operation string
	outcome   string
	duration  time.Duration
}

type fakePersistenceMetricsRecorder struct {
	calls []persistenceMetricCall
}

func (f *fakePersistenceMetricsRecorder) RecordHTTPRequest(context.Context, string, string, int, time.Duration) {}
func (f *fakePersistenceMetricsRecorder) RecordPaymentFlow(context.Context, string, string, time.Duration) {}
func (f *fakePersistenceMetricsRecorder) RecordPaymentCommand(context.Context, string, string, time.Duration) {}
func (f *fakePersistenceMetricsRecorder) RecordPersistenceOperation(_ context.Context, backend, resource, operation, outcome string, duration time.Duration) {
	f.calls = append(f.calls, persistenceMetricCall{backend: backend, resource: resource, operation: operation, outcome: outcome, duration: duration})
}

func TestStoreContractAgainstPostgres(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	store := NewStore(db, nil)
	contractStoreCRUD(t, store)
	contractStoreIdempotency(t, store)
	contractStoreSequence(t, store)
}

func TestUnitOfWorkContractAgainstPostgres(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	store := NewStore(db, nil)
	publisher := inprocess.NewPublisher()
	called := 0
	publisher.Subscribe("payment_intent.created", func(event domain.Event) error {
		called++
		return nil
	})

	uow := NewUnitOfWork(db, publisher, nil)
	if err := uow.Do(func(tx ports.Transaction) error {
		tx.SavePaymentIntent(domain.PaymentIntent{ID: "pi_1"})
		tx.SaveCharge(domain.Charge{ID: "ch_1", PaymentIntentID: "pi_1"})
		return tx.Publish(domain.PaymentIntentCreatedEvent{PaymentIntent: domain.PaymentIntent{ID: "pi_1"}})
	}); err != nil {
		t.Fatalf("do failed: %v", err)
	}

	if called != 1 {
		t.Fatalf("expected downstream publish once, got %d", called)
	}
	if _, err := store.GetPaymentIntent("pi_1"); err != nil {
		t.Fatalf("expected committed payment intent: %v", err)
	}
	if _, err := store.GetCharge("ch_1"); err != nil {
		t.Fatalf("expected committed charge: %v", err)
	}
	if count := queryInt(t, db, `SELECT count(*) FROM outbox_events WHERE event_name = 'payment_intent.created'`); count != 1 {
		t.Fatalf("expected 1 outbox row, got %d", count)
	}
}

func TestUnitOfWorkRollbackAgainstPostgres(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	uow := NewUnitOfWork(db, inprocess.NewPublisher(), nil)
	err := uow.Do(func(tx ports.Transaction) error {
		tx.SavePaymentIntent(domain.PaymentIntent{ID: "pi_1"})
		tx.SavePaymentAttempt(domain.PaymentAttempt{ID: "pa_1", PaymentIntentID: "pi_1"})
		_ = tx.Publish(domain.PaymentIntentCreatedEvent{PaymentIntent: domain.PaymentIntent{ID: "pi_1"}})
		return domain.NewError(500, "boom", "boom")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if count := queryInt(t, db, `SELECT count(*) FROM payment_intents WHERE id = 'pi_1'`); count != 0 {
		t.Fatalf("expected rollback of payment_intents, got %d", count)
	}
	if count := queryInt(t, db, `SELECT count(*) FROM payment_attempts WHERE id = 'pa_1'`); count != 0 {
		t.Fatalf("expected rollback of payment_attempts, got %d", count)
	}
	if count := queryInt(t, db, `SELECT count(*) FROM outbox_events WHERE event_name = 'payment_intent.created'`); count != 0 {
		t.Fatalf("expected rollback of outbox, got %d", count)
	}
}

func TestPersistenceMetricsAgainstPostgres(t *testing.T) {
	db, cleanup := openTestDB(t)
	defer cleanup()

	recorder := &fakePersistenceMetricsRecorder{}
	store := NewStore(db, recorder)
	store.SavePaymentIntent(domain.PaymentIntent{ID: "pi_1"})
	if _, err := store.GetPaymentIntent("pi_1"); err != nil {
		t.Fatalf("get payment intent failed: %v", err)
	}

	uow := NewUnitOfWork(db, inprocess.NewPublisher(), recorder)
	if err := uow.Do(func(tx ports.Transaction) error {
		tx.SavePaymentAttempt(domain.PaymentAttempt{ID: "pa_1", PaymentIntentID: "pi_1"})
		if _, err := tx.GetPaymentAttempt("pa_1"); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("uow do failed: %v", err)
	}

	want := []persistenceMetricCall{
		{backend: "postgres", resource: "payment_intent", operation: "save", outcome: "success"},
		{backend: "postgres", resource: "payment_intent", operation: "get", outcome: "success"},
		{backend: "postgres", resource: "payment_attempt", operation: "save", outcome: "success"},
		{backend: "postgres", resource: "payment_attempt", operation: "get", outcome: "success"},
	}

	if len(recorder.calls) != len(want) {
		t.Fatalf("expected %d calls, got %d: %#v", len(want), len(recorder.calls), recorder.calls)
	}
	for i := range want {
		got := recorder.calls[i]
		if got.backend != want[i].backend || got.resource != want[i].resource || got.operation != want[i].operation || got.outcome != want[i].outcome {
			t.Fatalf("call %d: expected %#v, got %#v", i, want[i], got)
		}
		if got.duration <= 0 {
			t.Fatalf("call %d: expected positive duration, got %s", i, got.duration)
		}
	}
}

func contractStoreCRUD(t *testing.T, store *Store) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	intent := domain.PaymentIntent{ID: "pi_1", MerchantID: "m_1", CustomerID: "c_1", Amount: 100, Currency: "USD", CaptureMethod: "manual", Status: domain.PaymentIntentRequiresPaymentMethod, CreatedAt: now, UpdatedAt: now}
	attempt := domain.PaymentAttempt{ID: "pa_1", PaymentIntentID: "pi_1", Status: domain.PaymentAttemptAuthorized, RequestedAt: now, RespondedAt: now}
	charge := domain.Charge{ID: "ch_1", PaymentIntentID: "pi_1", PaymentAttemptID: "pa_1", Amount: 100, CapturedAmount: 100, Status: domain.ChargeCaptured, CreatedAt: now, UpdatedAt: now}
	refund := domain.Refund{ID: "re_1", ChargeID: "ch_1", PaymentIntentID: "pi_1", Amount: 100, Status: domain.RefundSucceeded, CreatedAt: now, UpdatedAt: now}

	store.SavePaymentIntent(intent)
	store.SavePaymentAttempt(attempt)
	store.SaveCharge(charge)
	store.SaveRefund(refund)

	gotIntent, err := store.GetPaymentIntent("pi_1")
	if err != nil || gotIntent.ID != intent.ID || gotIntent.Amount != intent.Amount {
		t.Fatalf("unexpected intent: %+v err=%v", gotIntent, err)
	}
	gotAttempt, err := store.GetPaymentAttempt("pa_1")
	if err != nil || gotAttempt.ID != attempt.ID {
		t.Fatalf("unexpected attempt: %+v err=%v", gotAttempt, err)
	}
	gotCharge, err := store.GetCharge("ch_1")
	if err != nil || gotCharge.ID != charge.ID || gotCharge.CapturedAmount != charge.CapturedAmount {
		t.Fatalf("unexpected charge: %+v err=%v", gotCharge, err)
	}
	gotRefund, err := store.GetRefund("re_1")
	if err != nil || gotRefund.ID != refund.ID || gotRefund.Amount != refund.Amount {
		t.Fatalf("unexpected refund: %+v err=%v", gotRefund, err)
	}
}

func contractStoreIdempotency(t *testing.T, store *Store) {
	t.Helper()
	result, err := store.WithIdempotency("create:1", "fingerprint-1", func() (any, error) {
		return domain.PaymentIntent{ID: "pi_idem", Currency: "USD"}, nil
	})
	if err != nil {
		t.Fatalf("first idempotency call failed: %v", err)
	}
	if result.(domain.PaymentIntent).ID != "pi_idem" {
		t.Fatalf("unexpected idempotency result: %#v", result)
	}
	result, err = store.WithIdempotency("create:1", "fingerprint-1", func() (any, error) {
		return domain.PaymentIntent{ID: "pi_other", Currency: "USD"}, nil
	})
	if err != nil {
		t.Fatalf("second idempotency call failed: %v", err)
	}
	if result.(domain.PaymentIntent).ID != "pi_idem" {
		t.Fatalf("expected cached idempotency result, got %#v", result)
	}
	if _, err := store.WithIdempotency("create:1", "fingerprint-2", func() (any, error) {
		return domain.PaymentIntent{ID: "pi_conflict", Currency: "USD"}, nil
	}); err == nil {
		t.Fatal("expected idempotency conflict")
	}
}

func contractStoreSequence(t *testing.T, store *Store) {
	t.Helper()
	if got := store.NextID("pi"); got != "pi_000001" {
		t.Fatalf("unexpected first id: %s", got)
	}
	if got := store.NextReference("pr"); got != "pr_000002" {
		t.Fatalf("unexpected second ref: %s", got)
	}
}

func openTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err := Open(ctx, dsn)
		if err != nil {
			t.Fatalf("open db failed: %v", err)
		}
		if err := EnsureSchema(ctx, db); err != nil {
			_ = db.Close()
			t.Fatalf("ensure schema failed: %v", err)
		}
		return db, func() { _ = db.Close() }
	}

	container, db, err := startPostgresContainer(ctx)
	if err != nil {
		if isOptionalDockerError(err) {
			t.Skipf("skipping postgres contract tests: %v", err)
		}
		t.Fatalf("start postgres container failed: %v", err)
	}
	if err := EnsureSchema(ctx, db); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		t.Fatalf("ensure schema failed: %v", err)
	}
	return db, func() {
		_ = db.Close()
		_ = container.Terminate(ctx)
	}
}

func isOptionalDockerError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "Docker provider") || strings.Contains(msg, "rootless Docker not found") || strings.Contains(msg, "docker API")
}

func startPostgresContainer(ctx context.Context) (tc.Container, *sql.DB, error) {
	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image:        "postgres:17-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_DB":       "payment_sandbox",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		return nil, nil, err
	}
	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, err
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, err
	}
	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/payment_sandbox?sslmode=disable", host, port.Port())
	db, err := Open(ctx, dsn)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, err
	}
	return container, db, nil
}

func queryInt(t *testing.T, db *sql.DB, query string) int {
	t.Helper()
	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	return count
}
