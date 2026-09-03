# Webhook Retry Policy

## Goal

Define the simple synchronous retry policy used by the webhook dispatcher.

## Policy

- Max attempts: 3.
- This is the default limit enforced by the dispatcher.
- Backoff: linear, `attempt * 10ms`.
- Attempt 1 sends immediately.
- Attempt 2 sleeps `10ms` before retrying.
- Attempt 3 sleeps `20ms` before retrying.
- If the final attempt fails, return a typed delivery error with the `delivery_id` and attempt count.
- The dispatcher should retain a final delivery state for inspection.
- The dispatcher should expose a trace view with delivery metadata and attempt history for debug/reconciliation.

## Notes

- The policy is intentionally simple for the MVP.
- Sleep must remain injectable in tests.
- The payload bytes do not change across retries.
