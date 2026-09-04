# PaymentPlatformSDK Proposal

## Goal

Provide a small Ruby SDK for consuming `payment-sandbox` v1 from the Ruby app without coupling the app to raw HTTP calls.

## Positioning

- The SDK is a thin client, not a new source of truth.
- The OpenAPI `v1` spec remains canonical.
- The SDK should be a separate story from workflow consumer integration.

## Proposed Ruby API

```ruby
client = PaymentPlatform::Client.new(
  base_url: ENV.fetch("PAYMENT_PLATFORM_BASE_URL"),
  api_key: ENV.fetch("PAYMENT_PLATFORM_API_KEY"),
  tenant_id: ENV["PAYMENT_PLATFORM_TENANT_ID"],
  timeout: 5,
  retries: 2
)

client.create_payment_intent(amount: 1000, currency: "usd", capture_method: "manual")
client.confirm_payment_intent(id, payment_method_token: "pm_card_visa")
client.capture_payment_intent(id, amount: 1000)
client.create_refund(charge_id: "ch_123", amount: 1000)
client.get_payment_intent(id)
```

## Configuration

- `base_url`
- `api_key`
- `tenant_id` or equivalent account scope
- `timeout`
- `retries`
- automatic `Idempotency-Key` support for mutating calls

## Error Model

- Map API errors into a small set of Ruby exceptions.
- Keep provider-specific details accessible for debugging.
- Preserve status code, error code, and message.

## Contract Boundaries

- Requests and responses should mirror OpenAPI `v1`.
- Webhooks should be handled by a separate inbox/verifier layer.
- The SDK should not embed workflow-specific business logic.

## Recommended Next Step

- Implement `07-payment-platform-sdk` as a public, versioned Ruby gem.
