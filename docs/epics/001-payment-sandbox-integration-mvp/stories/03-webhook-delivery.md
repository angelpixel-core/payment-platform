---
id: 001-story-03-webhook-delivery
aliases: []
tags:
  - payments
  - sandbox
  - webhooks
epic: 001-payment-sandbox-integration-mvp
---

# Story: Webhook Delivery

## Intent

Simulate provider webhooks with retries, duplicates, and idempotent processing on the Rails side.

## Scope

### In Scope

- [ ] `payment.succeeded`
- [ ] `payment.failed`
- [ ] `payment.processing`
- [ ] Retry delivery
- [ ] Duplicate delivery
- [ ] Webhook inbox in Rails

### Out of Scope

- [ ] Complex signature standards beyond a simple shared secret
- [ ] External queue infrastructure

## System Design

```mermaid
flowchart LR
    API[Go Sandbox API] --> Q[Webhook Dispatcher]
    Q --> WH[Rails Webhook Endpoint]
    WH --> IN[Webhook Inbox]
    IN --> PM[Payment Projection]
    Q --> RT[Retry Scheduler]
```

## Explanation

1. The sandbox emits payment events after a lifecycle transition.
2. A dispatcher sends the event to Rails and tracks delivery attempts.
3. Rails stores every delivery in an inbox before applying business changes.
4. If the webhook fails, the dispatcher retries without breaking idempotency.
5. The local payment projection is updated only from validated inbox entries.

## Acceptance Criteria

- [ ] Rails can receive and store webhook deliveries.
- [ ] Duplicate events do not duplicate business effects.
- [ ] Failed deliveries are retried.

## Dependencies

- [ ] Sandbox API base
- [ ] Scenario engine
