package refunds

import (
	"strings"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type Service struct {
	clock  ports.Clock
	store  ports.Store
	events ports.EventPublisher
}

func NewService(store ports.Store, clock ports.Clock, events ports.EventPublisher) *Service {
	return &Service{store: store, clock: clock, events: events}
}

func (s *Service) CreateRefund(req domain.RefundRequest, idempotencyKey, fingerprint string) (domain.RefundResponse, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return domain.RefundResponse{}, domain.NewError(400, "missing_idempotency_key", "idempotency key is required")
	}
	if strings.TrimSpace(req.ChargeID) == "" {
		return domain.RefundResponse{}, domain.NewError(400, "invalid_charge_id", "charge_id is required")
	}

	key := "create_refund:" + idempotencyKey
	result, err := s.store.WithIdempotency(key, fingerprint, func() (any, error) {
		charge, err := s.charge(req.ChargeID)
		if err != nil {
			return nil, err
		}
		if charge.CapturedAmount == 0 {
			return nil, domain.NewError(409, "invalid_charge_state", "charge must be captured before refunding")
		}

		remaining := charge.CapturedAmount - charge.RefundedAmount
		amount := req.Amount
		if amount == 0 {
			amount = remaining
		}
		if amount <= 0 {
			return nil, domain.NewError(400, "invalid_amount", "refund amount must be greater than zero")
		}
		if amount > remaining {
			return nil, domain.NewError(400, "invalid_amount", "refund amount cannot exceed remaining captured amount")
		}

		now := s.clock.Now()
		refund := domain.Refund{ID: s.nextID("re"), ChargeID: charge.ID, PaymentIntentID: charge.PaymentIntentID, Amount: amount, Status: domain.RefundSucceeded, CreatedAt: now, UpdatedAt: now}
		charge.RefundedAmount += amount
		if charge.RefundedAmount == charge.CapturedAmount {
			charge.Status = domain.ChargeRefunded
		} else {
			charge.Status = domain.ChargePartiallyRefunded
		}
		charge.UpdatedAt = now
		s.store.SaveRefund(refund)
		s.store.SaveCharge(*charge)
		s.publish(domain.RefundCreatedEvent{Refund: refund, Charge: *charge})
		return domain.RefundResponse{Refund: refund, Charge: *charge}, nil
	})
	if err != nil {
		return domain.RefundResponse{}, err
	}
	return result.(domain.RefundResponse), nil
}

func (s *Service) charge(id string) (*domain.Charge, error) {
	charge, err := s.store.GetCharge(id)
	if err != nil {
		return nil, err
	}
	return &charge, nil
}

func (s *Service) nextID(prefix string) string { return s.store.NextID(prefix) }

func (s *Service) publish(event domain.Event) { if s.events != nil && event != nil { _ = s.events.Publish(event) } }
