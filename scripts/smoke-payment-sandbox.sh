#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:30001}"
run_id="$(date +%s)-$$"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

command -v curl >/dev/null || { printf 'curl is required\n' >&2; exit 1; }
command -v jq >/dev/null || { printf 'jq is required\n' >&2; exit 1; }

for _ in {1..30}; do
	if curl --silent --fail "$BASE_URL/health" >/dev/null 2>&1; then
		break
	fi
	sleep 1
done

request() {
  local method="$1" path="$2" key="$3" payload="${4:-}" output="$tmp_dir/response.json" status
  if [[ -n "$payload" ]]; then
    status="$(curl --fail-with-body --silent --show-error -o "$output" -w '%{http_code}' \
      -X "$method" "$BASE_URL$path" -H 'Accept: application/json' \
      -H 'Content-Type: application/json' -H "Idempotency-Key: $key" -H 'X-Sandbox-Scenario: approved_immediate' \
      --data "$payload")"
  else
    status="$(curl --fail-with-body --silent --show-error -o "$output" -w '%{http_code}' \
      -X "$method" "$BASE_URL$path" -H 'Accept: application/json' \
      -H "Idempotency-Key: $key")"
  fi
  printf '%s\n' "$status" > "$tmp_dir/status"
}

expect_status() {
  local expected="$1" actual
  actual="$(<"$tmp_dir/status")"
  if [[ "$actual" != "$expected" ]]; then
    printf 'Expected HTTP %s, got %s for %s\n%s\n' "$expected" "$actual" "$2" "$(<"$tmp_dir/response.json")" >&2
    exit 1
  fi
}

request GET /health health
expect_status 200 /health
[[ "$(jq -r '.status' "$tmp_dir/response.json")" == ok ]]

request POST /v1/payment_intents "smoke-create-$run_id" \
  "{\"amount\":1000,\"currency\":\"usd\",\"merchant_id\":\"merchant-smoke-$run_id\",\"customer_id\":\"customer-smoke-$run_id\",\"capture_method\":\"manual\"}"
expect_status 201 /v1/payment_intents
intent_id="$(jq -er '.payment_intent.id' "$tmp_dir/response.json")"

request POST "/v1/payment_intents/$intent_id/confirm" "smoke-confirm-$run_id" '{"payment_method_token":"pm_card_visa"}'
expect_status 200 confirm
[[ "$(jq -r '.payment_intent.status' "$tmp_dir/response.json")" == requires_capture ]]
attempt_id="$(jq -er '.payment_attempt.id' "$tmp_dir/response.json")"
charge_id="$(jq -er '.charge.id' "$tmp_dir/response.json")"
[[ "$(jq -r '.charge.payment_intent_id' "$tmp_dir/response.json")" == "$intent_id" ]]

request GET "/v1/payment_intents/$intent_id" get-intent
expect_status 200 get-intent
[[ "$(jq -r '.payment_intent.id' "$tmp_dir/response.json")" == "$intent_id" ]]
request GET "/v1/payment_attempts/$attempt_id" get-attempt
expect_status 200 get-attempt
request GET "/v1/charges/$charge_id" get-charge
expect_status 200 get-charge

request POST "/v1/payment_intents/$intent_id/capture" "smoke-capture-$run_id" '{}'
expect_status 200 capture
[[ "$(jq -r '.payment_intent.status' "$tmp_dir/response.json")" == succeeded ]]
[[ "$(jq -r '.charge.captured_amount' "$tmp_dir/response.json")" == 1000 ]]

request POST /v1/refunds "smoke-refund-$run_id" "{\"charge_id\":\"$charge_id\"}"
expect_status 201 refund
refund_id="$(jq -er '.refund.id' "$tmp_dir/response.json")"
[[ "$(jq -r '.charge.status' "$tmp_dir/response.json")" == refunded ]]

request GET "/v1/refunds/$refund_id" get-refund
expect_status 200 get-refund
request GET "/v1/payment_intents/$intent_id/lifecycle" lifecycle
expect_status 200 lifecycle
[[ "$(jq -r '.payment_lifecycle.status' "$tmp_dir/response.json")" == refunded ]]
[[ "$(jq -r '.payment_lifecycle.refundable_amount' "$tmp_dir/response.json")" == 0 ]]
[[ "$(jq -r '.payment_lifecycle.is_refundable' "$tmp_dir/response.json")" == false ]]

request GET /v1/reports/transactions report
expect_status 200 report
jq -e --arg id "$intent_id" '.transactions_report.transactions | any(.[]; (.payment_intent_id == $id or .payment_intent.id == $id))' "$tmp_dir/response.json" >/dev/null

request GET '/v1/reports/transactions?view=snapshot' snapshot
expect_status 200 snapshot
jq -e --arg id "$intent_id" '.transactions_snapshot.transactions | any(.[]; (.payment_intent_id == $id or .payment_intent.id == $id))' "$tmp_dir/response.json" >/dev/null

printf 'Smoke test passed for %s\n' "$intent_id"
