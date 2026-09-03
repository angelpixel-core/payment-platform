#!/usr/bin/env sh
set -eu

if command -v npx >/dev/null 2>&1; then
  npx @redocly/cli lint docs/openapi/payment-sandbox.v1.yaml
else
  echo "npx not available; install Node.js tooling to validate docs/openapi/payment-sandbox.v1.yaml" >&2
  exit 1
fi
