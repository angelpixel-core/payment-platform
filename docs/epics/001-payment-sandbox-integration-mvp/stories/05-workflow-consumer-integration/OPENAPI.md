# OpenAPI Guide

## Goal

Describe how to use the versioned OpenAPI contract for `payment-sandbox` v1.

## Files

- `docs/openapi/payment-sandbox.v1.yaml`
- `docs/openapi/payment-sandbox.v1.json`
- `docs/openapi/swagger-ui.html`
- `docs/openapi/redoc.html`

## Notes

- The YAML spec is the source of truth.
- The JSON export exists for tooling compatibility.
- Swagger UI and Redoc can coexist while evaluating the best browsing experience.

## Local Preview

Run the Go docs server from `apps/payment-sandbox`:

```bash
go run ./cmd/docs-server
```

Then open:

- `http://localhost:8081/`
- `http://localhost:8081/openapi/swagger-ui.html`
- `http://localhost:8081/openapi/redoc.html`
