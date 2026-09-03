# Workflow Consumer Pattern

## Pattern

Use an inbox-first consumer pattern.

## Steps

1. Receive webhook or fetch sandbox state.
2. Persist delivery/inbox entry immediately.
3. Validate signature and idempotency.
4. Apply business mutation only after validation.
5. Emit reconciliation snapshot if needed.

## Benefits

- Avoids duplicate business effects.
- Makes delivery/debug history queryable.
- Keeps the consumer decoupled from the sandbox implementation.
