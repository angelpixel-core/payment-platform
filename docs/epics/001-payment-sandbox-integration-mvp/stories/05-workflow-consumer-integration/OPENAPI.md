# OpenAPI Guide

## Goal

Describe how to use the versioned OpenAPI contract for `payment-sandbox` v1.

## Files

- `apps/payment-sandbox/docs/openapi/payment-sandbox.v1.yaml`
- `apps/payment-sandbox/docs/openapi/payment-sandbox.v1.json`
- `apps/payment-sandbox/docs/openapi/swagger-ui.html`
- `apps/payment-sandbox/docs/openapi/redoc.html`

## Notes

- The YAML spec is the source of truth.
- The JSON export is generated from the YAML and exists for tooling compatibility.
- From `apps/payment-sandbox`, run `go run ./cmd/openapi-export` to refresh the JSON mirror.
- Swagger UI and Redoc can coexist while evaluating the best browsing experience.

## Local Preview

Run the Go docs server from `apps/payment-sandbox`:

```bash
go run ./cmd/docs-server
```

Then open:

- `http://localhost:30001/openapi/swagger-ui.html`
- `http://localhost:30001/openapi/redoc.html`
