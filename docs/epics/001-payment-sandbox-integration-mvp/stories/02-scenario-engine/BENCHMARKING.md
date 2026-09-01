# Benchmarking and Profiling

## Goal

Keep lightweight performance checks available for `payment-sandbox` while the codebase evolves.

## Benchmarks

- `internal/adapters/inbound/http/middleware`:
  - `BenchmarkMetricsMiddleware`
  - `BenchmarkRequestIDMiddleware`
  - `BenchmarkObservabilityMiddleware`
  - `BenchmarkRecoveryPath`
- `internal/adapters/persistence/memory`:
  - `BenchmarkMemoryStoreSaveGetPaymentIntent`
  - `BenchmarkMemoryUnitOfWorkDo`
  - `BenchmarkMemoryOutboxPublish`
- `internal/adapters/messaging/outbox`:
  - `BenchmarkOutboxPublish`
- `internal/sandbox`:
  - `BenchmarkCreatePaymentIntent`
  - `BenchmarkPaymentLifecycle`

## Recommended Commands

```bash
go test ./...
go test -run=^$ -bench=. -benchmem ./...
go test -race ./...
go test -run=^$ -bench=BenchmarkCreatePaymentIntent -cpuprofile cpu.out -memprofile mem.out ./...
```

## Notes

- Use the non-integration benchmarks for fast feedback.
- Add Postgres integration benchmarks only if the DB path becomes a performance target.
