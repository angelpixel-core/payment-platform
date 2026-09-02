package domain

import "testing"

func TestDomainEventVersionsAreStable(t *testing.T) {
	tests := []struct {
		name    string
		event   Event
		wantVer int
	}{
		{name: "created", event: PaymentIntentCreatedEvent{}, wantVer: EventVersionV1},
		{name: "confirmed", event: PaymentIntentConfirmedEvent{}, wantVer: EventVersionV1},
		{name: "finalized", event: PaymentIntentFinalizedEvent{}, wantVer: EventVersionV1},
		{name: "captured", event: PaymentIntentCapturedEvent{}, wantVer: EventVersionV1},
		{name: "refund created", event: RefundCreatedEvent{}, wantVer: EventVersionV1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.EventVersion(); got != tt.wantVer {
				t.Fatalf("expected version %d, got %d", tt.wantVer, got)
			}
		})
	}
}
