# vegeta load-pressure scaffold

This directory contains reproducible HTTP pressure targets for `payment-sandbox`.

## Targets

- `health.txt` hits `GET /health` for a lightweight smoke check.
- `reports-transactions.txt` hits `GET /v1/reports/transactions` for a read-path pressure test.

## Usage

```bash
BASE_URL=http://localhost:30001 \
TARGETS_FILE=apps/payment-sandbox/load-tests/vegeta/reports-transactions.txt \
DURATION=30s \
RATE=50 \
apps/payment-sandbox/load-tests/vegeta/pressure.sh
```

## Environment Variables

- `BASE_URL` defaults to `http://localhost:30001`.
- `TARGETS_FILE` defaults to `reports-transactions.txt`.
- `DURATION` defaults to `30s`.
- `RATE` defaults to `50`.
- `WORKERS` defaults to `10`.

## Notes

- This is a pressure scaffold, not a business-flow runner.
- Use `k6` for scripted multi-step scenarios.
