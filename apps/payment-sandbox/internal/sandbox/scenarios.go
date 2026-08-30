package sandbox

import "payment-sandbox/internal/domain"

type ScenarioName = domain.ScenarioName
type ScenarioOutcome = domain.ScenarioOutcome

const (
	ScenarioApprovedImmediate         ScenarioName = domain.ScenarioApprovedImmediate
	ScenarioDeclinedInsufficientFunds ScenarioName = domain.ScenarioDeclinedInsufficientFunds
	ScenarioRequiresAction3DS         ScenarioName = domain.ScenarioRequiresAction3DS
	ScenarioProcessingThenSucceeded   ScenarioName = domain.ScenarioProcessingThenSucceeded
)

func normalizeScenarioName(value string) ScenarioName {
	return domain.NormalizeScenarioName(value)
}
