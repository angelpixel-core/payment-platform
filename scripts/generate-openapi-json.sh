#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/../apps/payment-sandbox"
go run ./cmd/openapi-export
