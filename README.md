# Payment Platform

## Test Targets

```bash
make test-unit
make test-integration
make test-contract
make test-load-business
make test-load-pressure
```

## Local Stack

The local development stack will be defined at the project root and operated through these Make targets:

```bash
make stack/up
make stack/seed
make stack/down
```

Use `make stack/down` followed by `make stack/up` to restart the stack. A reset target may remove local volumes when a clean database is required.

The sandbox is available from the host at `http://localhost:8080`. Services inside Docker Compose should reach it using the service hostname `payment-sandbox`.

Local configuration will be documented in `.env.example`; production secrets must not be committed.
