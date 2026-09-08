# Workflow Consumer Contract

## Goal

Define the generic consumer-side contract for integrating any workflow owner with `payment-sandbox` v1.

## Purpose

- Keep the consumer decoupled from sandbox internals.
- Make payment state, webhooks, and reconciliation inspectable locally.
- Reconcile against the sandbox transaction report and snapshot export without mutating business state.
- Preserve a narrow, stable `v1` boundary for future Rails workflows.

## Rules

- The consumer owns its local state.
- The sandbox owns payment truth.
- Webhooks are stored before business mutation.
- Idempotency is enforced by the consumer.
- Contract changes must go through `v1`.
- Reconciliation is read-only.
- Snapshot comparison must be deterministic and diff-friendly.

## Consumer Shape

- The consumer is any workflow owner with a local payment projection.
- The consumer should depend on a small gateway interface and webhook inbox.
- The consumer should be able to reconcile its local projection against `GET /v1/reports/transactions` and `GET /v1/reports/transactions?view=snapshot`.

## Gateway Interface

The consumer gateway should expose a small set of operations:

- create payment intent
- confirm payment intent
- capture payment intent
- refund charge
- fetch payment intent status
- fetch transaction report
- fetch reconciliation snapshot

The gateway should return sandbox errors in a normalized consumer shape, not raw transport errors.

## Local State

The consumer should persist at least:

- payment intents
- payment attempts
- charges
- refunds
- webhook inbox entries
- idempotency keys
- reconciliation snapshots

## Data Expectations

- Payment intent lifecycle is created, confirmed, captured, refunded, and queried through the sandbox API.
- Webhook deliveries carry a stable `delivery_id` and `event_id`.
- Reconciliation snapshots must be comparable to sandbox reports.
- Local records should retain enough metadata to debug duplicate deliveries and reconciliation mismatches.

## Error Model

The consumer should map sandbox failures into stable domain-level errors such as:

- invalid request
- payment declined
- payment pending
- payment not found
- duplicate request
- reconciliation mismatch

## Reconciliation Contract

Reconciliation should:

- compare local projection state against sandbox report data
- use `GET /v1/reports/transactions` for the canonical current report
- use `GET /v1/reports/transactions?view=snapshot` when a stable export is needed for diffing
- record mismatch context without changing business truth
- remain queryable by payment intent identifier
