# Webhook Projection Contract

## Goal

Define the Rails local payment projection that is updated from validated inbox entries and read by reconciliation without mutating business state.

## Purpose

- Turn verified inbox deliveries into a stable local projection.
- Keep the projection idempotent and queryable.
- Give reconciliation a consistent read model to compare against sandbox truth.

## Projection Rules

1. Only validated inbox entries may update the projection.
2. Duplicate inbox entries must not duplicate business effects.
3. Out-of-order deliveries must not regress the projection.
4. Reconciliation reads the projection but does not mutate it.
5. The projection must remain inspectable for debugging and audit.

## Inputs

The projection consumes inbox entries that are already:

- persisted
- signature-validated
- deduplicated or marked as duplicates
- associated with a stable `delivery_id` and `event_id`

## Suggested Projection States

- `pending` - the payment exists locally but has not been confirmed by a validated delivery.
- `confirmed` - the inbox delivery validated a successful payment transition.
- `failed` - the inbox delivery validated a terminal failure.
- `processing` - the payment is in an intermediate state.
- `refunded` - the payment or charge has been refunded.
- `partially_refunded` - only part of the captured amount has been refunded.

## Update Behavior

- `payment.succeeded` should move the projection toward `confirmed` or `refunded` depending on later refund state.
- `payment.failed` should move the projection toward `failed`.
- `payment.processing` should move the projection toward `processing` until a later validated transition arrives.
- If an inbox record is marked duplicate, the projection must not change.
- If a later delivery represents a newer state, it may advance the projection but must not rewind it.

## Reconciliation Read Contract

Reconciliation may read:

- payment intent status
- latest attempt status
- charge status
- refunded amount
- delivery metadata for the last applied inbox entry

Reconciliation must:

- compare the local projection with sandbox truth
- report mismatches
- never mutate the projection directly

## Query Expectations

Operators should be able to inspect:

- the current projection state
- the inbox record that produced it
- the latest delivery identifiers
- the last applied event type

## Notes

- This contract is Rails-side only.
- It complements `WEBHOOK_INBOX_CONTRACT.md`.
- It intentionally stops short of defining the reconciliation job implementation.
