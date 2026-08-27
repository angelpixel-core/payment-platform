---
id: 001-story-02-scenario-engine
aliases: []
tags:
  - payments
  - sandbox
  - scenarios
epic: 001-payment-sandbox-integration-mvp
---

# Story: Scenario Engine

## Intent

Make the sandbox deterministic by driving payment outcomes from named scenarios.

## Scope

### In Scope

- [ ] `approved_immediate`
- [ ] `declined_insufficient_funds`
- [ ] `requires_action_3ds`
- [ ] `processing_then_succeeded`
- [ ] Scenario selection via header or payment method token

### Out of Scope

- [ ] Real fraud or risk scoring
- [ ] Crypto scenarios
- [ ] Provider-specific edge cases beyond the shared card flow

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

## Acceptance Criteria

- [ ] The sandbox returns the expected status for each named scenario.
- [ ] Scenario selection is deterministic and reproducible.
- [ ] Pending flows can transition to a final state later.

## Dependencies

- [ ] Sandbox API base
