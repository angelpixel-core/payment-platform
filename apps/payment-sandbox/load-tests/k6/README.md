# k6 business-sequence scaffold

This directory contains the initial `k6` script for deterministic payment flows against `payment-sandbox`.

## Script

- `business-sequences.js` creates a payment intent, confirms it with a sandbox scenario, and optionally captures it.

## Usage

```bash
BASE_URL=http://localhost:30001/v1 \
SCENARIO=approved_immediate \
CAPTURE_METHOD=manual \
k6 run apps/payment-sandbox/load-tests/k6/business-sequences.js
```

## Environment Variables

- `BASE_URL` defaults to `http://localhost:30001/v1`.
- `SCENARIO` defaults to `approved_immediate`.
- `CAPTURE_METHOD` defaults to `manual`.
- `AMOUNT`, `CURRENCY`, `MERCHANT_ID`, `CUSTOMER_ID`, `PAYMENT_METHOD_TOKEN` override the request payload.
- `VUS`, `DURATION`, and `SLEEP` control the run shape.

## Notes

- This is a scaffold, not a full benchmark suite.
- `k6` is the fit for scripted business flows; `vegeta` remains the fit for raw HTTP pressure.
