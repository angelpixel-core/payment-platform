package scenarios

import (
	"fmt"
	"strings"

	"payment-sandbox/internal/domain"
)

type ScenarioEngine struct {
	config scenarioConfig
}

func NewScenarioEngine() *ScenarioEngine { return &ScenarioEngine{config: defaultScenarioConfig()} }

func (e *ScenarioEngine) Resolve(headerScenario, paymentMethodToken string) (domain.ScenarioName, error) {
	if scenario := domain.NormalizeScenarioName(headerScenario); scenario != "" {
		if e.config.isKnownScenario(scenario) {
			return scenario, nil
		}
		return "", domain.NewError(422, "invalid_scenario", fmt.Sprintf("unknown sandbox scenario %q", headerScenario))
	}

	if token := strings.TrimSpace(paymentMethodToken); token != "" {
		if scenario, ok := e.config.tokenToScenario[token]; ok {
			return scenario, nil
		}
		return "", domain.NewError(422, "invalid_scenario", fmt.Sprintf("unknown payment method token %q", paymentMethodToken))
	}

	return domain.ScenarioApprovedImmediate, nil
}

func (e *ScenarioEngine) Outcome(name domain.ScenarioName) (domain.ScenarioOutcome, error) {
	switch domain.NormalizeScenarioName(string(name)) {
	case domain.ScenarioApprovedImmediate:
		return domain.ScenarioOutcome{Scenario: domain.ScenarioApprovedImmediate, IntentStatus: domain.PaymentIntentRequiresCapture, AttemptStatus: domain.PaymentAttemptAuthorized, ChargeStatus: domain.ChargeAuthorized, CreatesCharge: true}, nil
	case domain.ScenarioDeclinedInsufficientFunds:
		return domain.ScenarioOutcome{Scenario: domain.ScenarioDeclinedInsufficientFunds, IntentStatus: domain.PaymentIntentFailed, AttemptStatus: domain.PaymentAttemptDeclined, DeclineCode: "insufficient_funds"}, nil
	case domain.ScenarioRequiresAction3DS:
		return domain.ScenarioOutcome{Scenario: domain.ScenarioRequiresAction3DS, IntentStatus: domain.PaymentIntentRequiresAction, AttemptStatus: domain.PaymentAttemptRequiresAction}, nil
	case domain.ScenarioProcessingThenSucceeded:
		return domain.ScenarioOutcome{Scenario: domain.ScenarioProcessingThenSucceeded, IntentStatus: domain.PaymentIntentProcessing, AttemptStatus: domain.PaymentAttemptSubmitted, FinalizesLater: true}, nil
	default:
		return domain.ScenarioOutcome{}, domain.NewError(422, "invalid_scenario", fmt.Sprintf("unknown sandbox scenario %q", name))
	}
}
