package sandbox

import (
	"fmt"
	"strings"
)

type ScenarioConfig struct {
	tokenToScenario map[string]ScenarioName
}

func DefaultScenarioConfig() ScenarioConfig {
	return ScenarioConfig{
		tokenToScenario: map[string]ScenarioName{
			"pm_card_visa":                    ScenarioApprovedImmediate,
			"pm_card_insufficient_funds":      ScenarioDeclinedInsufficientFunds,
			"pm_card_authentication_required": ScenarioRequiresAction3DS,
			"pm_card_processing":              ScenarioProcessingThenSucceeded,
		},
	}
}

func (c ScenarioConfig) Resolve(headerScenario, paymentMethodToken string) (ScenarioName, error) {
	if scenario := normalizeScenarioName(headerScenario); scenario != "" {
		if c.isKnownScenario(scenario) {
			return scenario, nil
		}
		return "", newError(422, "invalid_scenario", fmt.Sprintf("unknown sandbox scenario %q", headerScenario))
	}

	if token := strings.TrimSpace(paymentMethodToken); token != "" {
		if scenario, ok := c.tokenToScenario[token]; ok {
			return scenario, nil
		}
		return "", newError(422, "invalid_scenario", fmt.Sprintf("unknown payment method token %q", paymentMethodToken))
	}

	return ScenarioApprovedImmediate, nil
}

func (c ScenarioConfig) Outcome(name ScenarioName) (ScenarioOutcome, error) {
	switch normalizeScenarioName(string(name)) {
	case ScenarioApprovedImmediate:
		return ScenarioOutcome{
			Scenario:      ScenarioApprovedImmediate,
			IntentStatus:  PaymentIntentRequiresCapture,
			AttemptStatus: PaymentAttemptAuthorized,
			ChargeStatus:  ChargeAuthorized,
			CreatesCharge: true,
		}, nil
	case ScenarioDeclinedInsufficientFunds:
		return ScenarioOutcome{
			Scenario:      ScenarioDeclinedInsufficientFunds,
			IntentStatus:  PaymentIntentFailed,
			AttemptStatus: PaymentAttemptDeclined,
			DeclineCode:   "insufficient_funds",
		}, nil
	case ScenarioRequiresAction3DS:
		return ScenarioOutcome{
			Scenario:      ScenarioRequiresAction3DS,
			IntentStatus:  PaymentIntentRequiresAction,
			AttemptStatus: PaymentAttemptRequiresAction,
		}, nil
	case ScenarioProcessingThenSucceeded:
		return ScenarioOutcome{
			Scenario:       ScenarioProcessingThenSucceeded,
			IntentStatus:   PaymentIntentProcessing,
			AttemptStatus:  PaymentAttemptSubmitted,
			FinalizesLater: true,
		}, nil
	default:
		return ScenarioOutcome{}, newError(422, "invalid_scenario", fmt.Sprintf("unknown sandbox scenario %q", name))
	}
}

func (c ScenarioConfig) isKnownScenario(name ScenarioName) bool {
	switch normalizeScenarioName(string(name)) {
	case ScenarioApprovedImmediate, ScenarioDeclinedInsufficientFunds, ScenarioRequiresAction3DS, ScenarioProcessingThenSucceeded:
		return true
	default:
		return false
	}
}
