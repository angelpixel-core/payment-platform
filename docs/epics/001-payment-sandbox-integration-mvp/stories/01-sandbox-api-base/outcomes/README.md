# Story 01 Outcomes

This folder contains the manual request sequence for the Sandbox API Base story.

## Shared Variables

- `BASE_URL`: sandbox base URL, usually `http://localhost:8080`
- `IDEMPOTENCY_KEY`: unique key for each operation; reuse it to verify idempotency
- `PAYMENT_INTENT_ID`: id returned by the create request
- `CHARGE_ID`: id returned by confirm or capture, used for the refund request

## Execution Order

1. `01-create-payment-intent.http`
2. `02-confirm-payment-intent.http`
3. `03-capture-payment-intent.http`
4. `04-refund-payment-intent.http`

## Manual Execution

- Start the sandbox with `cd apps/payment-sandbox && go run ./cmd/payment-sandbox`.
- Open each `*.http` file in order with a REST client that supports `.http` files, such as VS Code REST Client.
- Send `01-create-payment-intent.http` first.
- Copy the `payment_intent.id` from step 1 into steps 2 and 3.
- Copy the `charge.id` from step 2 or step 3 into step 4.
- Reuse the same `IDEMPOTENCY_KEY` only when validating idempotency.
- Change `IDEMPOTENCY_KEY` when you want a fresh operation.

## Expected Result

- Step 1 creates a payment intent in `requires_payment_method`.
- Step 2 confirms the intent and creates an authorized charge.
- Step 3 captures the charge and moves the intent to `succeeded`.
- Step 4 refunds the captured charge.
