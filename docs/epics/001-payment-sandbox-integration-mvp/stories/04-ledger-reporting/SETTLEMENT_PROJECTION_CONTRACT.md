# Settlement Projection Contract

## Goal

Define a daily settlement projection grouped by `merchant_id`, `currency`, and `settlement_date` so Rails can see the batches that should be settled from liquidable funds.

## Purpose

- Model settlement as a daily batch, not a per-cargo ad hoc adjustment.
- Keep captured manual charges traceable inside a batch.
- Make refunds visible in the same batch net amount.

## Grouping

- `merchant_id`
- `currency`
- `settlement_date`

## Rules

1. Only captured manual-capture charges participate in settlement batches.
2. The settlement date comes from the charge capture timestamp in UTC.
3. Refunds reduce the same batch net amount.
4. Automatic-capture charges stay in `available` balance, not in settlement batches.
5. The projection is read-only and derived from the payment records.

## Response Shape

```json
{
  "generated_at": "2026-09-04T00:00:00Z",
  "count": 1,
  "batches": [
    {
      "merchant_id": "merchant_1",
      "currency": "usd",
      "settlement_date": "2026-09-04",
      "status": "pending",
      "gross_amount": 300,
      "refunded_amount": 0,
      "net_amount": 300,
      "charge_count": 1,
      "refund_count": 0,
      "charge_ids": ["ch_1"],
      "charges": [
        {
          "charge_id": "ch_1",
          "payment_intent_id": "pi_1",
          "captured_at": "2026-09-04T00:00:00Z",
          "amount": 300,
          "refunded_amount": 0
        }
      ]
    }
  ]
}
```

## Notes

- This is a projection, not a persisted settlement ledger.
- It is enough to model daily payout batches and reconcile against liquidable balances.
- If settlement persistence is added later, this shape can become the read model for it.
