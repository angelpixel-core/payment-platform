# Balance Projection Contract

## Goal

Define a read-only balance projection grouped by `merchant_id`, `currency`, and `account_type` so Rails can reconcile merchant balances against sandbox truth.

## Purpose

- Separate merchant balances into operational buckets.
- Keep the projection simple enough for reconciliation.
- Make the buckets explicit and stable.

## Grouping

- `merchant_id`
- `currency`
- `account_type`

## Account Types

- `available` - captured funds that are already considered available.
- `reserved` - funds reserved by authorized or processing payments that have not been captured yet.
- `liquidable` - captured funds that are ready to be liquidated but are not yet considered available.

## Rules

1. Reserved balances come from payment intents that are still awaiting capture or finalization.
2. Liquidable balances come from captured charges on manual-capture flows.
3. Available balances come from captured charges on automatic-capture flows.
4. Refunds reduce the same bucket that originally held the captured amount.
5. The projection must remain read-only and derivable from the existing payment lifecycle records.
6. Manual-capture settlement batches are derived from the `liquidable` bucket and grouped by capture day.

## Response Shape

```json
{
  "generated_at": "2026-09-04T00:00:00Z",
  "count": 3,
  "balances": [
    {
      "merchant_id": "merchant_1",
      "currency": "usd",
      "account_type": "available",
      "amount": 200,
      "updated_at": "2026-09-04T00:00:00Z"
    }
  ]
}
```

## Notes

- The projection is intentionally simpler than a full accounting ledger.
- It is good enough for reconciliation and merchant balance inspection.
- The daily settlement projection consumes the `liquidable` bucket without changing this shape.
