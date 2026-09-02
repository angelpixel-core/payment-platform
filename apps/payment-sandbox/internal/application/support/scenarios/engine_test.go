package scenarios

import (
	"testing"

	"payment-sandbox/internal/domain"
)

func TestScenarioResolution(t *testing.T) {
	engine := NewScenarioEngine()

	tests := []struct {
		name           string
		headerScenario string
		token          string
		wantScenario   domain.ScenarioName
		wantError      bool
	}{
		{name: "header priority", headerScenario: "declined_insufficient_funds", token: "pm_card_visa", wantScenario: domain.ScenarioDeclinedInsufficientFunds},
		{name: "token fallback", headerScenario: "", token: "pm_card_insufficient_funds", wantScenario: domain.ScenarioDeclinedInsufficientFunds},
		{name: "unknown header", headerScenario: "unknown_scenario", token: "pm_card_visa", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario, err := engine.Resolve(tt.headerScenario, tt.token)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected invalid scenario error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve failed: %v", err)
			}
			if scenario != tt.wantScenario {
				t.Fatalf("expected %s, got %s", tt.wantScenario, scenario)
			}
		})
	}
}
