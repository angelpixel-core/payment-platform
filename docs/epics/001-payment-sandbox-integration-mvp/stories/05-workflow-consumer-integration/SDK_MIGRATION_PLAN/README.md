# Payment SDK Ruby Migration Plan

## Objective

Move the Ruby consumer reference artifacts into an executable, reusable SDK that consumes the Payment API contract without requiring its own server.

The SDK will run in the local Ruby interpreter or as a gem dependency. It will communicate with a configurable Payment API endpoint and will not encode infrastructure environments such as Sandbox, Test, or Prod in its namespace.

## Product Structure

Payment Platform is the product. `Payment` groups the product artifacts:

```text
apps/
└── payment/
    ├── service/
    │   └── Go payment service
    └── sdk/
        └── ruby/
            └── Ruby SDK
```

The current `payment-sandbox` name is a historical implementation name. It describes an environment, not the product or the SDK namespace. The service and module rename will happen incrementally after the SDK migration.

## Namespace

The Ruby SDK namespace uses the acronym `SDK`:

```ruby
Payment::SDK::Client
Payment::SDK::Payloads
Payment::SDK::Signatures
Payment::SDK::Webhooks
Payment::SDK::Inbox
Payment::SDK::Reconciliation
Payment::SDK::Errors
```

The following names are intentionally not used:

```ruby
Payment::Sandbox
Payment::SandboxSDK
Payment::Sdk
```

## Environment Configuration

The SDK receives endpoint and contract configuration from its consumer:

```ruby
Payment::SDK::Client.new(
  base_url: ENV.fetch("PAYMENT_API_BASE_URL"),
  api_version: "v1",
  signing_secret: ENV["PAYMENT_SIGNING_SECRET"]
)
```

Environment names are configuration concerns, not namespaces. The current local endpoint is:

```text
PAYMENT_API_BASE_URL=http://localhost:10201
```

## Target SDK Layout

```text
apps/payment/sdk/ruby/
├── Gemfile
├── payment-sdk.gemspec
├── README.md
├── lib/
│   ├── payment-sdk.rb
│   └── payment/
│       └── sdk/
│           ├── client.rb
│           ├── errors.rb
│           ├── payloads.rb
│           ├── signatures.rb
│           ├── webhooks.rb
│           ├── inbox.rb
│           └── reconciliation.rb
└── test/
```

## Artifact Mapping

```text
docs/.../workflow_inbox.rb
  -> apps/payment/sdk/ruby/lib/payment/sdk/inbox.rb

docs/.../workflow_reconciliation.rb
  -> apps/payment/sdk/ruby/lib/payment/sdk/reconciliation.rb

docs/.../workflow_inbox_test.rb
  -> apps/payment/sdk/ruby/test/payment/sdk/inbox_test.rb

docs/.../workflow_reconciliation_test.rb
  -> apps/payment/sdk/ruby/test/payment/sdk/reconciliation_test.rb
```

The existing files remain the reference baseline until their SDK equivalents have equivalent or stronger tests.

## Responsibilities

### Client

- [ ] Consume the versioned Payment API over HTTP.
- [ ] Create, confirm, capture, and refund payment intents.
- [ ] Query payment entities, lifecycle, reports, and snapshots.
- [ ] Expose a small consumer-facing interface.

### Payloads and Errors

- [ ] Build payloads matching OpenAPI v1.
- [ ] Normalize API errors into stable SDK errors.
- [ ] Preserve response and request correlation metadata.

### Signatures and Webhooks

- [ ] Verify webhook signatures.
- [ ] Enforce timestamp tolerance and constant-time comparison.
- [ ] Parse webhook payloads into SDK event objects.

### Inbox and Reconciliation

- [ ] Preserve persist-before-mutate inbox ordering.
- [ ] Prevent duplicate delivery side effects.
- [ ] Compare local projections with Payment reports and snapshots.
- [ ] Produce inspectable reconciliation mismatches.

## Migration Checklist

### Foundation

- [ ] Create `apps/payment/service` as the target service location.
- [ ] Create `apps/payment/sdk/ruby`.
- [ ] Add the Ruby gem structure and local test command.
- [ ] Define the `Payment::SDK` namespace.

### SDK Implementation

- [ ] Implement the HTTP client.
- [ ] Implement OpenAPI v1 payloads.
- [ ] Implement SDK error mapping.
- [ ] Implement idempotency-key handling.
- [ ] Implement signature verification.
- [ ] Migrate inbox processing.
- [ ] Migrate reconciliation.
- [ ] Migrate and expand tests.

### Integration

- [ ] Add local configuration using `PAYMENT_API_BASE_URL`.
- [ ] Run SDK tests with the local Payment service.
- [ ] Validate the SDK against the local smoke flow.
- [ ] Document Ruby interpreter and Bundler commands.
- [ ] Link the SDK evidence from the story checklist.

### Cleanup and Renaming

- [ ] Retire the executable Ruby artifacts from the story directory after migration.
- [ ] Update story 05 to reference the SDK implementation.
- [ ] Rename the Go service directory from `payment-sandbox` to `payment/service`.
- [ ] Rename the Go module and imports when external references are ready.
- [ ] Update Docker, Compose, and documentation names.

## Migration Order

1. Create the SDK structure without changing the existing service.
2. Migrate the contract, payloads, errors, and signatures.
3. Migrate inbox and reconciliation behavior.
4. Add equivalent tests and local API integration tests.
5. Update story evidence and consumer documentation.
6. Retire the old Ruby reference artifacts.
7. Rename the Go service from `payment-sandbox` to `payment/service` as a separate controlled change.

## Architectural Constraints

- [ ] The SDK must not start an HTTP server.
- [ ] The SDK must not import or depend on Go service code.
- [ ] The SDK must communicate through the versioned Payment API contract.
- [ ] `Sandbox`, `Test`, and `Prod` must remain endpoint/configuration concepts.
- [ ] `Payment::SDK` must remain independent from Rails and any specific consumer application.
