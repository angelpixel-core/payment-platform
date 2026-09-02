# Webhook Delivery Contract

## Goal

Define a single webhook envelope for all sandbox payment events so retries, idempotency, and Rails inbox processing share one stable contract.

## Supported Events

- `payment.succeeded`
- `payment.failed`
- `payment.processing`

## Envelope

All webhook deliveries use the same JSON body shape.

```json
{
  "schema_version": 1,
  "event_type": "payment.succeeded",
  "event_id": "evt_01J...",
  "delivery_id": "del_01J...",
  "attempt": 1,
  "occurred_at": "2026-09-02T12:34:56Z",
  "payment_intent_id": "pi_01J...",
  "data": {}
}
```

## Common Fields

- `schema_version`: contract version for backward-compatible evolution.
- `event_type`: one of the supported event names.
- `event_id`: stable identifier for the business event.
- `delivery_id`: stable identifier for the delivery attempt group.
- `attempt`: 1-based retry counter for the same `delivery_id`.
- `occurred_at`: UTC timestamp of the lifecycle transition.
- `payment_intent_id`: payment intent associated with the event.
- `data`: event-specific payload.

## Event Data

### `payment.succeeded`

```json
{
  "charge_id": "ch_01J...",
  "amount": 1099,
  "currency": "usd"
}
```

### `payment.failed`

```json
{
  "failure_code": "insufficient_funds",
  "failure_message": "card was declined"
}
```

### `payment.processing`

```json
{
  "next_action": "processing_then_succeeded"
}
```

## Signature

- The HTTP request must include `X-Sandbox-Signature`.
- The signature is computed over the exact request body bytes with a shared secret.
- Rails must reject deliveries with missing or invalid signatures before any state change.

## Delivery Rules

- Retries reuse the same body.
- Only `attempt` changes across retries.
- Duplicate `delivery_id` values must be treated as the same delivery group.
- Rails inbox entries must be able to store the original payload and the verified delivery metadata.

## Notes

- This contract is intentionally narrow.
- Additional provider-specific webhook formats are out of scope for this story.
- The envelope can evolve through `schema_version` without changing the top-level shape.
