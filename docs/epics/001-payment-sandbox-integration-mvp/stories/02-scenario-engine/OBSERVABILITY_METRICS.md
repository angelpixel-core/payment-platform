# Observability Metrics Contract

## Goal

Define the metrics surface for `payment-sandbox` unit of work execution before wiring the implementation.

## Unit of Work Metrics

### Recorder API

```go
func (r *Recorder) RecordUnitOfWork(ctx context.Context, backend, outcome string, duration time.Duration)
```

### Labels

- `uow.backend`: `memory` | `postgres`
- `uow.outcome`: `success` | `rollback` | `commit_error`

### Metric Names

- `payment_sandbox_unit_of_work_total`
- `payment_sandbox_unit_of_work_duration_ms`
- `payment_sandbox_unit_of_work_errors_total`

## Outcome Rules

- `success` when `Do(...)` completes without error.
- `rollback` when the callback returns an error and the transaction is not committed.
- `commit_error` when commit or post-commit publish fails.

## Backend Rules

- `memory` for `internal/adapters/persistence/memory/uow.go`.
- `postgres` for `internal/adapters/persistence/postgres/uow.go`.

## Implementation Notes

- Measure the full `Do(...)` duration, not just the callback body.
- Keep the recorder optional so existing tests and non-observability paths keep working.
- Prefer a shared recorder method over separate metrics helpers per backend.

## Outbox Metrics

### Recorder API

```go
func (r *Recorder) RecordOutboxOperation(ctx context.Context, backend, operation, outcome string, duration time.Duration)
func (r *Recorder) RecordOutboxPending(ctx context.Context, backend string, pending int64)
```

### Labels

- `outbox.backend`: `memory` | `postgres`
- `outbox.operation`: `enqueue` | `publish`
- `outbox.outcome`: `success` | `failure`

### Metric Names

- `payment_sandbox_outbox_operations_total`
- `payment_sandbox_outbox_operation_duration_ms`
- `payment_sandbox_outbox_operation_errors_total`
- `payment_sandbox_outbox_pending_events`

### Outcome Rules

- `enqueue` records the event entering the outbox queue/table.
- `publish` records the downstream dispatch attempt.
- `failure` is used when downstream publish fails.
- `pending` is the current number of queued/unpublished records after the operation.

### Backend Rules

- `memory` for `internal/adapters/messaging/outbox/publisher.go`.
- `postgres` for the transaction outbox row lifecycle in `internal/adapters/persistence/postgres/{store.go,uow.go}`.

### Implementation Notes

- Measure enqueue and publish duration separately.
- Keep pending counts aligned with the outbox history/queue state, not just successful publishes.
