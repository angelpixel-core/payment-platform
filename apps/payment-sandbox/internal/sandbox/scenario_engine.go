package sandbox

import "payment-sandbox/internal/modules/scenarios"

type ScenarioEngine = scenarios.Engine

func NewScenarioEngine() *ScenarioEngine { return scenarios.NewEngine() }
