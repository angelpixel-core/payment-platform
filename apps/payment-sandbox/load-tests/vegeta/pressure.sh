#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:10201}"
TARGETS_FILE="${TARGETS_FILE:-$(dirname "$0")/reports-transactions.txt}"
DURATION="${DURATION:-30s}"
RATE="${RATE:-50}"
WORKERS="${WORKERS:-10}"

tmp_targets="$(mktemp)"
trap 'rm -f "$tmp_targets"' EXIT

sed "s|http://localhost:8080|$BASE_URL|g" "$TARGETS_FILE" >"$tmp_targets"

vegeta attack -duration "$DURATION" -rate "$RATE" -workers "$WORKERS" -targets "$tmp_targets" | vegeta report
