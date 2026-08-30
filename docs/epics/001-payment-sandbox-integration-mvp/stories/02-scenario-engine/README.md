---
id: 001-story-02-scenario-engine
aliases: []
tags:
  - payments
  - sandbox
  - scenarios
epic: 001-payment-sandbox-integration-mvp
status: done
reason:
---

# Story: Scenario Engine

## Intent

Make the sandbox deterministic by driving payment outcomes from named scenarios.

## Scope

### In Scope

- [x] `approved_immediate`
- [x] `declined_insufficient_funds`
- [x] `requires_action_3ds`
- [x] `processing_then_succeeded`
- [x] Scenario selection via header or payment method token
- [x] Deterministic mapping from scenario input to outcome

## Implementation Checklist

- [x] Create `internal/sandbox/scenarios.go` for scenario resolution.
- [x] Create `internal/sandbox/scenario_config.go` for allowed scenarios and token mapping.
- [x] Extend `internal/server/server.go` to read `X-Sandbox-Scenario`.
- [x] Integrate the scenario engine into `ConfirmPaymentIntent`.
- [x] Ensure header selection has priority over `payment_method_token`.
- [x] Return a stable error when no scenario can be resolved.
- [x] Add tests for header priority, token fallback, and invalid scenarios.
- [x] Add tests for each outcome: approved, declined, processing, requires action.
- [x] Update documentation status when the story is complete.

### Out of Scope

- [ ] Real fraud or risk scoring
- [ ] Crypto scenarios
- [ ] Provider-specific edge cases beyond the shared card flow
- [ ] Webhook delivery
- [ ] Ledger projections

## System Design

```mermaid
flowchart LR
    Rails[Rails App] --> API[Go Sandbox API]
    API --> SE[Scenario Engine]
    SE --> S1[approved]
    SE --> S2[declined]
    SE --> S3[processing]
    SE --> S4[requires_action]
    SE --> DB[(Scenario Config)]
```

## Explanation

1. Rails passes a scenario marker with the payment request.
2. The scenario engine maps that marker to a deterministic outcome.
3. The engine may emit immediate success, rejection, or delayed processing.
4. Scenario config stays external so new flows can be added without changing the core API.
5. This keeps payment behavior reproducible across development, demos, and tests.

## Contract Notes

- Scenario selection should accept either a header or a payment method token.
- The same scenario input must always produce the same outcome.
- Scenario definitions should be externalized in config, not hardcoded in controllers.
- Unknown scenarios should fail fast with a stable error payload.

## Acceptance Criteria

- [x] The sandbox returns the expected status for each named scenario.
- [x] Scenario selection is deterministic and reproducible.
- [x] Pending flows can transition to a final state later.
- [x] Unknown scenario names are rejected clearly.
- [x] Scenario config can be changed without modifying the API contract.

## Dependencies

- [ ] Sandbox API base

## Related Docs

- [Level 2 Module Boundaries](./LEVEL2_MODULE_BOUNDARIES.md)
- [Implementation Checklist](./IMPLEMENTATION_CHECKLIST.md)
- [Refactor Proposal](./REFACTOR_PROPOSAL.md)
- [Architecture Plan](./ARCHITECTURE_PLAN.md)
- [Roadmap](./ROADMAP.md)
