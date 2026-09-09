# OpenAPI Guide

## Goal

Describe how to use the versioned OpenAPI contract for `payment-sandbox` v1.

## Files

- `apps/payment-sandbox/docs/openapi/v1/payment-sandbox.v1.yaml`
- `apps/payment-sandbox/docs/openapi/v1/payment-sandbox.v1.json`
- `apps/payment-sandbox/docs/swagger/index.html`
- `apps/payment-sandbox/docs/redoc/index.html`

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

- `http://localhost:10201/openapi/swagger-ui.html`
- `http://localhost:10201/openapi/redoc.html`
