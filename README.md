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
make stack/up ARGS="--build -d"
make stack/seed
make stack/smoke
make stack/down
make stack/reset
```

Use `make stack/down` followed by `make stack/up ARGS="-d"` to restart the stack. Pass Compose options through `ARGS`, for example `make stack/up ARGS="--build -d"` for a force-build and detached startup. `make stack/reset` removes local volumes when a clean database is required.

The sandbox API and its OpenAPI documentation are available from the host at `http://localhost:30001`. Services inside Docker Compose should reach the API using the service hostname `payment-sandbox`.

Local configuration will be documented in `.env.example`; production secrets must not be committed.
