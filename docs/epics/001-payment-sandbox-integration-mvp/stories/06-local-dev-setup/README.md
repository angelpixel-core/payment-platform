---
id: 001-story-06-local-dev-setup
aliases: []
tags:
  - devops
  - docker
  - sandbox
epic: 001-payment-sandbox-integration-mvp
status: pending
---

# Story: Local Dev Setup

## Intent

Make the sandbox and Rails app runnable locally with Docker Compose, with the API and its live documentation exposed by the same HTTP process.

## Setup Decision

- `docker-compose.yml`, `.env.example`, and the project-level development README belong at the repository root.
- Stack operations use namespaced Make targets: `make stack/up`, `make stack/down`, and `make stack/seed`.
- The Rails application is documented as an external consumer because it is not part of this repository.
- The host-facing API and documentation URL is `http://localhost:30001`; Compose services use `http://payment-sandbox:8080`.
- PostgreSQL is exposed to the host at `localhost:30002` and remains available inside Compose at `postgres:5432`.
- Swagger UI, Redoc, and the OpenAPI documents are HTTP routes of the sandbox process, not a separate service.

## Implementation Order

1. Define the Compose services and health dependencies.
2. Add reproducible local environment variables.
3. Serve API and OpenAPI documentation from the same HTTP process.
4. Add deterministic seeds through the existing `payment-sandbox seed` command.
5. Add smoke tests for health, payment lifecycle, report, and snapshot.
6. Add Make targets for startup, shutdown, seeding, smoke testing, and reset.

## Scope

### In Scope

- [x] Docker Compose services ([evidence](../../../../../../docker-compose.yml))
- [x] Local PostgreSQL ([evidence](../../../../../../docker-compose.yml))
- [x] Local scenario seeds ([evidence](../../../../../../apps/payment-sandbox/internal/application/operations/seed.go))
- [x] Smoke test scripts ([evidence](../../../../../../scripts/smoke-payment-sandbox.sh))
- [x] One-command startup for the full stack ([evidence](../../../../../../Makefile))
- [x] Local env config for Rails and Go ([evidence](../../../../../../.env.example))

### Out of Scope

- [ ] Cloud deployment
- [ ] Production observability stack
- [ ] Kubernetes or staging infrastructure
- [ ] CI/CD pipeline automation

## System Design

```mermaid
flowchart LR
    Dev[Developer] --> HTTP[Sandbox HTTP :30001]
    HTTP --> API[JSON API /v1]
    HTTP --> Docs[Swagger Redoc OpenAPI]
    HTTP --> PG[(PostgreSQL :30002)]
    Dev --> RT[Seed + Smoke Tests]
    Rails[Rails App] --> API
```

## Explanation

1. Docker Compose starts PostgreSQL and the sandbox process with its HTTP adapter.
2. The HTTP adapter exposes JSON API routes and live OpenAPI documentation from one process and origin.
3. Rails consumes the API through the same host or Compose network URL.
4. Seed data provides deterministic test scenarios.
5. Smoke tests verify the basic payment flow before feature work begins.
6. This setup keeps the MVP reproducible for development and demos.

## Contract Notes

- The full stack should start with a single local command.
- Rails and Go must use predictable local hostnames and ports.
- API and documentation must share the same HTTP origin.
- Seed data should include at least one happy-path and one failure-path scenario.
- Smoke tests should prove the basic payment loop end to end.
- The setup should be simple enough to reset from scratch.

## Acceptance Criteria

- [x] The full stack runs with one local command. ([evidence](../../../../../../Makefile))
- [x] Seeded scenarios are available. ([evidence](../../../../../../apps/payment-sandbox/internal/application/operations/seed.go))
- [ ] Smoke tests validate the happy path.
- [x] Local config is documented and reproducible. ([evidence](../../../../../../.env.example))
- [x] The developer can reset the environment without manual cleanup. ([evidence](../../../../../../Makefile))

## Dependencies

- [x] Sandbox API base
- [ ] Rails adapter (external to this repository)
- [x] Scenario engine
- [x] Ledger and reporting
