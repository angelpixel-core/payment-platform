package ports

import "payment-sandbox/internal/domain"

type ScenarioResolver interface {
	Resolve(headerScenario, paymentMethodToken string) (domain.ScenarioName, error)
	Outcome(name domain.ScenarioName) (domain.ScenarioOutcome, error)
}
