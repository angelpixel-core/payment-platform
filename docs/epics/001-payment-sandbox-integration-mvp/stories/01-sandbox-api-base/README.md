---
id: 001-story-01-sandbox-api-base
aliases: []
tags:
  - payments
  - sandbox
  - api
epic: 001-payment-sandbox-integration-mvp
---

# Story: Sandbox API Base

## Intent

Implement the core payment API that Rails uses to create, confirm, capture, and refund card payments against the sandbox.

## Scope

### In Scope

- [ ] `POST /v1/payment_intents`
- [ ] `POST /v1/payment_intents/:id/confirm`
- [ ] `POST /v1/payment_intents/:id/capture`
- [ ] `POST /v1/refunds`
- [ ] Domain models for `PaymentIntent`, `PaymentAttempt`, `Charge`, and `Refund`
- [ ] Validations for amount, currency, state transitions, and required references
- [ ] Basic idempotency for create, confirm, and refund requests

### Out of Scope

- [ ] Webhook delivery
- [ ] Reporting and reconciliation
- [ ] Rails adapter implementation
- [ ] Scenario selection logic
- [ ] Ledger projections

## System Design

```mermaid
flowchart LR
    Rails[Rails App] --> API[Go Sandbox API]
    API --> PI[PaymentIntent]
    API --> PA[PaymentAttempt]
    API --> CH[Charge]
    API --> RF[Refund]
    API --> DB[(PostgreSQL)]
```

## Explanation

1. Rails calls the sandbox through a narrow payment API instead of talking to domain tables directly.
2. `PaymentIntent` holds the commercial goal, while `PaymentAttempt` captures each technical try.
3. `Charge` represents the financial obligation created by a successful or authorized attempt.
4. `Refund` references the original charge and tracks reversals independently.
5. PostgreSQL is the source of truth for this story, and the API stays thin so later stories can add scenarios, webhooks, and reporting without breaking the contract.

## Contract Notes

- `PaymentIntent` should move through explicit states such as `requires_payment_method`, `requires_confirmation`, `processing`, `requires_capture`, `succeeded`, `cancelled`, and `failed`.
- `PaymentAttempt` should record the processor response, decline reason, and timestamps.
- `Refund` should support full and partial amounts, always linked to the original charge.
- Invalid transitions must fail fast with a stable error payload.

## Acceptance Criteria

- [ ] Rails can create and progress a payment intent.
- [ ] Refunds are registered and linked to the original charge.
- [ ] Invalid transitions are rejected.
- [ ] Duplicate create/confirm/refund requests do not create duplicate financial effects.
- [ ] The API returns predictable error payloads for invalid state transitions.

## Dependencies

- [ ] Shared domain model for payment lifecycle
