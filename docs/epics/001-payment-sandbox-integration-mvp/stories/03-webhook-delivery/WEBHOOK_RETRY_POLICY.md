# Webhook Retry Policy

## Goal

Define the simple synchronous retry policy used by the webhook dispatcher.

## Policy

- Max attempts: 3.
- Backoff: linear, `attempt * 10ms`.
- Attempt 1 sends immediately.
- Attempt 2 sleeps `10ms` before retrying.
- Attempt 3 sleeps `20ms` before retrying.
- If the final attempt fails, return the last transport error.

## Notes

- The policy is intentionally simple for the MVP.
- Sleep must remain injectable in tests.
- The payload bytes do not change across retries.
