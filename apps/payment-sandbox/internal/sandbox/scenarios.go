package sandbox

import "strings"

type ScenarioName string

const (
	ScenarioApprovedImmediate         ScenarioName = "approved_immediate"
	ScenarioDeclinedInsufficientFunds ScenarioName = "declined_insufficient_funds"
	ScenarioRequiresAction3DS         ScenarioName = "requires_action_3ds"
	ScenarioProcessingThenSucceeded   ScenarioName = "processing_then_succeeded"
)

type ScenarioOutcome struct {
	Scenario       ScenarioName
	IntentStatus   PaymentIntentStatus
	AttemptStatus  PaymentAttemptStatus
	ChargeStatus   ChargeStatus
	DeclineCode    string
	CreatesCharge  bool
	FinalizesLater bool
}

func normalizeScenarioName(value string) ScenarioName {
	return ScenarioName(strings.ToLower(strings.TrimSpace(value)))
}
