# Local Dev Setup Implementation Checklist

## Architecture

- [x] Consolidate `docs-server` into the sandbox HTTP process. [Evidence](../../../../../../apps/payment-sandbox/internal/adapters/inbound/http/docs/handler.go)
- [x] Serve Swagger UI from the sandbox process. [Evidence](../../../../../../apps/payment-sandbox/internal/adapters/inbound/http/server.go)
- [x] Serve Redoc from the sandbox process. [Evidence](../../../../../../apps/payment-sandbox/internal/adapters/inbound/http/server.go)
- [x] Serve OpenAPI YAML and JSON from the sandbox process. [Evidence](../../../../../../apps/payment-sandbox/internal/adapters/inbound/http/server.go)
- [x] Keep `/v1/*` and `/openapi/*` under the same HTTP origin. [Evidence](../../../../../../apps/payment-sandbox/internal/adapters/inbound/http/server.go)
- [ ] Keep transport concerns in `internal/adapters/inbound/http`.
- [ ] Keep application and domain layers independent from HTTP and HTML.

## Ports

- [x] Expose API and documentation on host port `10201`. [Evidence](../../../../../../docker-compose.yml)
- [x] Expose PostgreSQL on host port `10202`. [Evidence](../../../../../../docker-compose.yml)
- [x] Keep API internal container port at `8080`.
- [x] Keep PostgreSQL internal container port at `5432`.
- [ ] Remove operational host URL references to ports `8080` and `8081`.
- [x] Update load-test and smoke-test defaults to port `10201`.

## Docker Compose

- [x] Define PostgreSQL service.
- [x] Define payment sandbox service.
- [x] Serve documentation from the sandbox container and process. [Evidence](../../../../../../apps/payment-sandbox/Dockerfile)
- [x] Add host port variables for `10201` and `10202` to `.env.example`. [Evidence](../../../../../../.env.example)
- [x] Verify `/health` at `http://localhost:10201/health`.

## Seeds

- [x] Add deterministic local scenario fixtures.
- [x] Use stable fixture IDs and timestamps.
- [x] Persist seeds through the configured unit of work.
- [x] Make repeated seed execution idempotent.

## Smoke Tests

- [x] Add Bash, `curl`, and `jq` smoke script.
- [x] Validate required local dependencies.
- [x] Wait for service readiness before requests.
- [ ] Validate health successfully on port `10201`.
- [ ] Validate create, confirm, capture, and refund.
- [ ] Validate lifecycle, report, and snapshot.
- [ ] Resolve the PostgreSQL `current transaction is aborted` failure.
- [ ] Mark the happy-path acceptance criterion after the full flow passes.

## Documentation

- [x] Document root-level stack commands.
- [ ] Document the unified API and documentation origin.
- [x] Document Swagger at `http://localhost:10201/openapi/swagger-ui.html`.
- [ ] Update OpenAPI examples to use port `30001`.
- [ ] Link implementation evidence from the story README.

## Acceptance Criteria

- [ ] API and Swagger work from the same port.
- [ ] Stack starts with explicit arguments, for example `make stack/up ARGS="--build -d"`.
- [x] Seeds work against a clean database.
- [ ] The complete smoke test passes.
- [x] The environment can be reset with `make stack/reset`.
