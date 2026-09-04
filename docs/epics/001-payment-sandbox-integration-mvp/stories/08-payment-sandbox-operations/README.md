---
id: 001-story-08-payment-sandbox-operations
aliases: []
tags:
  - payments
  - sandbox
  - cli
  - load-testing
  - observability
epic: 001-payment-sandbox-integration-mvp
status: pending
---

# Story: Payment Sandbox Operations

## Intent

Add operational tooling around `payment-sandbox` so the team can replay scenarios, generate load, inspect metrics, and run local infrastructure with the same domain vocabulary.

## Scope

### In Scope

- [ ] Cobra-based CLI subcommands under `payment-sandbox`.
- [ ] Scenario runner commands for deterministic payment flows.
- [ ] Load-test helpers for sequential and parallel execution.
- [ ] Technical metrics for requests, latency, errors, CPU, memory, outbox, and webhooks.
- [ ] Business dashboards for intents, captures, refunds, webhook outcomes, settlement, and reconciliation.
- [ ] Docker Compose for sandbox, database, telemetry, and dashboard services.
- [ ] Make targets for stack up/down with optional detached mode and volume cleanup.

### Out of Scope

- [ ] Rails business logic implementation.
- [ ] SDK packaging beyond the separate gem story.
- [ ] Production deployment orchestration.
- [ ] External queue providers.

## Artifact Names

- `sandbox` - the Go API/service runtime.
- `database` - the PostgreSQL service.
- `telemetry` - Prometheus.
- `dashboard` - Grafana.
- `scenario-runner` - CLI subcommands for deterministic flows.
- `webhooks` - webhook delivery/debugging views.
- `reconciliation` - reconciliation views and reports.

## CLI Shape

```bash
payment-sandbox simulate <scenario>
payment-sandbox replay <scenario>
payment-sandbox burst --count 100 --parallel 20
payment-sandbox seed
```

## Load Testing Position

- `k6` for business-sequence scenarios.
- `vegeta` for reproducible HTTP pressure.
- Go goroutines only when a domain-specific runner is needed.

## Observability Position

- Prometheus scrapes application metrics.
- Grafana renders technical and business dashboards.
- The application publishes counters, histograms, and gauges; it does not talk directly to dashboards.

## Runtime Position

- Development uses PostgreSQL by default.
- Tests use the in-memory adapter.
- Docker Compose should include the sandbox, PostgreSQL, Prometheus, and Grafana.

## Related Docs

- [Implementation Checklist](./IMPLEMENTATION_CHECKLIST.md)
- [Story 02: Scenario Engine](../02-scenario-engine/README.md)
- [Story 04: Ledger Reporting](../04-ledger-reporting/README.md)
- [Story 05: Workflow Consumer Integration](../05-workflow-consumer-integration/README.md)
- [Story 07: Payment Platform SDK](../07-payment-platform-sdk/README.md)
