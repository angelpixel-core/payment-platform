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

## Consumed Endpoints

The consumer should consume the following `v1` endpoints:

- `POST /v1/payment_intents`
- `POST /v1/payment_intents/{id}/confirm`
- `POST /v1/payment_intents/{id}/capture`
- `POST /v1/refunds`
- `GET /v1/payment_intents/{id}`
- `GET /v1/payment_intents/{id}/lifecycle`
- `GET /v1/reports/transactions`
- `GET /v1/reports/transactions?view=snapshot`

The inspection endpoints are read-only helpers for debugging and local projection visibility. The report endpoints are the reconciliation source of truth.

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

### Error Mapping

Suggested mapping from sandbox codes to consumer errors:

- `invalid_amount` -> `invalid_request`
- `invalid_currency` -> `invalid_request`
- `missing_idempotency_key` -> `invalid_request`
- `invalid_intent_state` -> `invalid_request`
- `invalid_attempt_state` -> `invalid_request`
- `invalid_charge_state` -> `invalid_request`
- `payment_intent_not_found` -> `payment_not_found`
- `payment_attempt_not_found` -> `payment_not_found`
- `charge_not_found` -> `payment_not_found`
- `refund_not_found` -> `payment_not_found`
- `idempotency_conflict` -> `duplicate_request`
- `invalid_scenario` -> `payment_declined` when triggered by business scenario resolution, otherwise `invalid_request`
- `internal_error` -> `upstream_error`
- `panic` -> `upstream_error`

The consumer may preserve the original sandbox code as debug metadata, but the public consumer contract should expose only the normalized error shape.

## Reconciliation Contract

Reconciliation should:

- compare local projection state against sandbox report data
- use `GET /v1/reports/transactions` for the canonical current report
- use `GET /v1/reports/transactions?view=snapshot` when a stable export is needed for diffing
- record mismatch context without changing business truth
- remain queryable by payment intent identifier
