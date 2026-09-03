# Webhook Signature Configuration

## Goal

Define how both sides obtain the shared secret used to verify sandbox webhook deliveries.

## Decision

- Use `WEBHOOK_SIGNING_SECRET` as the shared configuration key on both Go and Rails.
- Store it as an environment variable in both runtimes.
- Keep the value out of payloads, docs examples, and logs.

## Go Sandbox

- Read `WEBHOOK_SIGNING_SECRET` from the process environment.
- Fail fast at startup if the variable is missing or empty.
- Use the secret to sign the raw webhook body before delivery.

## Rails

- Read `WEBHOOK_SIGNING_SECRET` from the process environment.
- Reject deliveries when the secret is missing or the signature cannot be verified.
- Verify the signature before writing business state.

## Verification Rule

- The signature is computed against the exact request body bytes.
- Retries must reuse the same body so the signature stays stable for the same payload.

## Notes

- This story intentionally uses one shared environment variable name on both sides.
- If the secret is missing in local or CI, the delivery flow must fail loudly instead of guessing a default.
