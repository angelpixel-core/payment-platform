package scenarios

import "payment-sandbox/internal/domain"

type Config struct {
	tokenToScenario map[string]domain.ScenarioName
}

func DefaultConfig() Config {
	return Config{
		tokenToScenario: map[string]domain.ScenarioName{
			"pm_card_visa":                    domain.ScenarioApprovedImmediate,
			"pm_card_insufficient_funds":      domain.ScenarioDeclinedInsufficientFunds,
			"pm_card_authentication_required": domain.ScenarioRequiresAction3DS,
			"pm_card_processing":              domain.ScenarioProcessingThenSucceeded,
		},
	}
}

func (c Config) isKnownScenario(name domain.ScenarioName) bool {
	switch domain.NormalizeScenarioName(string(name)) {
	case domain.ScenarioApprovedImmediate, domain.ScenarioDeclinedInsufficientFunds, domain.ScenarioRequiresAction3DS, domain.ScenarioProcessingThenSucceeded:
		return true
	default:
		return false
	}
}
