# Payment Sandbox

Local card payment sandbox for the payment-platform docs.

## Run

```bash
go test ./...
go run ./cmd/payment-sandbox
```

## Common Targets

```bash
make test-unit
make test-integration
make test-contract
make test-load-business
make test-load-pressure
```
