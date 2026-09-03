# Workflow Consumer Contract

## Goal

Define the consumer-side contract for integrating a workflow owner with `payment-sandbox` v1.

## Rules

- The consumer owns its local state.
- The sandbox owns payment truth.
- Webhooks are stored before business mutation.
- Idempotency is enforced by the consumer.
- Contract changes must go through `v1`.

## Consumer Shape

- The consumer may be `OrderProcessor` or any workflow owner with a local payment projection.
- The consumer should depend on a small gateway interface and webhook inbox.

## Data Expectations

- Payment intent lifecycle is created, confirmed, captured, refunded, and queried through the sandbox API.
- Webhook deliveries carry a stable `delivery_id` and `event_id`.
- Reconciliation snapshots must be comparable to sandbox reports.
