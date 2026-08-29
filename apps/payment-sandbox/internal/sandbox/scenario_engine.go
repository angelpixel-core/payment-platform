package sandbox

import (
	"fmt"
	"strings"
)

type ScenarioEngine struct {
	config ScenarioConfig
}

func NewScenarioEngine() *ScenarioEngine {
	return &ScenarioEngine{config: DefaultScenarioConfig()}
}

func (e *ScenarioEngine) Resolve(headerScenario, paymentMethodToken string) (ScenarioName, error) {
	if scenario := normalizeScenarioName(headerScenario); scenario != "" {
		if e.config.isKnownScenario(scenario) {
			return scenario, nil
		}
		return "", newError(422, "invalid_scenario", fmt.Sprintf("unknown sandbox scenario %q", headerScenario))
	}

	if token := strings.TrimSpace(paymentMethodToken); token != "" {
		if scenario, ok := e.config.tokenToScenario[token]; ok {
			return scenario, nil
		}
		return "", newError(422, "invalid_scenario", fmt.Sprintf("unknown payment method token %q", paymentMethodToken))
	}

	return ScenarioApprovedImmediate, nil
}

func (e *ScenarioEngine) Outcome(name ScenarioName) (ScenarioOutcome, error) {
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
