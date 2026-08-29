package sandbox

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

func (c ScenarioConfig) isKnownScenario(name ScenarioName) bool {
	switch normalizeScenarioName(string(name)) {
	case ScenarioApprovedImmediate, ScenarioDeclinedInsufficientFunds, ScenarioRequiresAction3DS, ScenarioProcessingThenSucceeded:
		return true
	default:
		return false
	}
}
