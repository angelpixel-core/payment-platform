# Transaction Report Contract

## Goal

Define the `GET /v1/reports/transactions` payload so Rails can reconcile local payment projections against sandbox truth.

## Purpose

- Expose a stable read model for reconciliation.
- Surface payment, attempt, charge, and refund state together.
- Keep the report traceable back to the underlying payment lifecycle records.
- Include a balance projection grouped by `merchant_id`, `currency`, and `account_type`.
- Include a daily settlement projection grouped by `merchant_id`, `currency`, and `settlement_date`.

## Endpoint

- `GET /v1/reports/transactions`

## Response Shape

```json
{
  "transactions_report": {
    "generated_at": "2026-09-04T00:00:00Z",
    "count": 1,
    "balance_projection": {
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
    },
    "settlement_projection": {
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
    },
    "transactions": [
      {
        "payment_intent": {
          "id": "pi_1",
          "status": "succeeded"
        },
        "latest_attempt": {
          "id": "pa_1",
          "status": "authorized"
        },
        "charge": {
          "id": "ch_1",
          "status": "refunded"
        },
        "refunds": [
          {
            "id": "re_1",
            "status": "succeeded"
          }
        ]
      }
    ]
  }
}
```

## Rules

1. One report line corresponds to one payment intent projection.
2. The report must include enough nested state for reconciliation.
3. The report must remain read-only.
4. Refunds are attached to their payment intent and charge context.
5. The report should be stable enough to diff against Rails snapshots.
6. The balance projection must expose `available`, `reserved`, and `liquidable` buckets per merchant and currency.
7. The settlement projection must include daily manual-capture batches and net refunds against each batch.

## Notes

- This contract is intentionally smaller than a full ledger engine.
- It is enough for reconciliation and reporting consumers.
- Later ledger expansion can add derived balances without breaking this payload shape.
