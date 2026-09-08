---
id: CONTRACT
aliases: []
tags: []
---

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
- `Idempotency-Key` is required on every mutating request and must be sent as a header.
- No body fallback is allowed for idempotency keys in the base contract.
- `v1` is the only supported contract version for this story.
- Any backward-incompatible request, response, or error-shape change requires a new contract version, not a silent change to `v1`.
- Reconciliation is read-only.
- Snapshot comparison must be deterministic and diff-friendly.

## Consumer Shape

- The consumer is any workflow owner with a local payment projection.
- The consumer should depend on a small gateway interface and webhook inbox.
- The consumer should be able to reconcile its local projection against `GET /v1/reports/transactions` and `GET /v1/reports/transactions?view=snapshot`.
- All mutating gateway operations must send `Idempotency-Key`.

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

## Inbox Model

The consumer should model webhook persistence as `payment_webhook_inbox` composed on top of a generic `webhook_inbox` boundary.

The inbox entry should include at least:

- `delivery_id`
- `event_id`
- `event_type`
- `payload`
- `status`
- `received_at`
- `processed_at`

Suggested inbox statuses:

- `received`
- `validated`
- `processed`
- `duplicate`
- `failed`

The inbox should persist before any business mutation, and duplicate `delivery_id` values must not reapply side effects.

## Traceability

The inbox should retain enough information to trace each delivery end-to-end:

- `delivery_id` and `event_id` must remain queryable.
- The inbox should keep a reference to the affected payment record when mutation succeeds.
- Duplicate, failed, and rejected deliveries must remain visible in the inbox history.
- Delivery state should be inspectable independently from business projection state.
- The consumer should be able to answer which delivery caused which local mutation.

## Deduplication Rules

- The consumer should dedupe primarily by `delivery_id`.
- `event_id` should be retained for traceability and may be used as an additional protection against accidental event replays.
- A repeated `delivery_id` must never reapply business effects.
- A repeated `event_id` with a new `delivery_id` should be treated as a replay unless an explicit manual override exists.
- Intentional reprocessing must require an explicit replay action or override flag.

## Idempotency Rules

- All mutating operations must include a non-empty `Idempotency-Key` header.
- The consumer should persist the key before applying side effects.
- A repeated request with the same key and payload must resolve to the original result.
- A repeated request with the same key and different payload must fail with a stable duplicate/idempotency conflict error.
- Read-only inspection and reconciliation endpoints do not require idempotency keys.

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
- `payment pending` and `reconciliation mismatch` are consumer-side states or errors, not sandbox transport codes.

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

## Mismatch Reporting

- `match` results should be persisted but should not be reported as mismatches.
- `mismatch`, `missing_local`, `missing_remote`, and `stale` results should be sent to a separate mismatch reporter.
- Reporting must not mutate payment projections, inbox entries, or sandbox state.
- Snapshots must be persisted before reporting so a reporter failure does not lose reconciliation evidence.
- Reprocessing the same `run_id` must not create duplicate snapshot or mismatch records.
