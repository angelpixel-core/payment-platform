---
id: 001-story-06-local-dev-setup
aliases: []
tags:
  - devops
  - docker
  - sandbox
epic: 001-payment-sandbox-integration-mvp
---

# Story: Local Dev Setup

## Intent

Make the sandbox and Rails app runnable locally with Docker Compose.

## Scope

### In Scope

- [ ] Docker Compose services
- [ ] Local PostgreSQL
- [ ] Local scenario seeds
- [ ] Smoke test scripts
- [ ] One-command startup for the full stack
- [ ] Local env config for Rails and Go

### Out of Scope

- [ ] Cloud deployment
- [ ] Production observability stack
- [ ] Kubernetes or staging infrastructure
- [ ] CI/CD pipeline automation

## System Design

```mermaid
flowchart LR
    Dev[Developer] --> DC[Docker Compose]
    DC --> Rails[Rails App]
    DC --> Go[Go Sandbox]
    DC --> PG[(PostgreSQL)]
    DC --> RT[Seed + Smoke Tests]
```

## Explanation

1. Docker Compose starts the local environment with all required services.
2. Rails and the Go sandbox share the same local network and database dependencies.
3. Seed data provides deterministic test scenarios.
4. Smoke tests verify the basic payment flow before feature work begins.
5. This setup keeps the MVP reproducible for development and demos.
6. Local config should make it easy to switch between fake and sandbox providers.

## Contract Notes

- The full stack should start with a single local command.
- Rails and Go must use predictable local hostnames and ports.
- Seed data should include at least one happy-path and one failure-path scenario.
- Smoke tests should prove the basic payment loop end to end.
- The setup should be simple enough to reset from scratch.

## Acceptance Criteria

- [ ] The full stack runs with one local command.
- [ ] Seeded scenarios are available.
- [ ] Smoke tests validate the happy path.
- [ ] Local config is documented and reproducible.
- [ ] The developer can reset the environment without manual cleanup.

## Dependencies

- [ ] Sandbox API base
- [ ] Rails adapter
- [ ] Scenario engine
- [ ] Ledger and reporting
