.PHONY: test test-unit test-integration test-contract test-system test-load-business test-load-pressure

test: test-unit

test-unit:
	cd apps/payment-sandbox && go test ./...

test-integration:
	cd apps/payment-sandbox && go test ./internal/adapters/persistence/postgres ./internal/adapters/messaging/webhook ./internal/adapters/inbound/http/... ./internal/server ./internal/docs/openapi

test-contract:
	./scripts/validate-openapi.sh

test-system:
	@echo "system tests are not defined yet"

test-load-business:
	BASE_URL=$${BASE_URL:-http://localhost:8080/v1} SCENARIO=$${SCENARIO:-approved_immediate} k6 run apps/payment-sandbox/load-tests/k6/business-sequences.js

test-load-pressure:
	BASE_URL=$${BASE_URL:-http://localhost:8080} TARGETS_FILE=$${TARGETS_FILE:-apps/payment-sandbox/load-tests/vegeta/reports-transactions.txt} DURATION=$${DURATION:-30s} RATE=$${RATE:-50} apps/payment-sandbox/load-tests/vegeta/pressure.sh
