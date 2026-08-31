package refunds

import (
	"strings"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type Service struct {
	uow   ports.UnitOfWork
	clock ports.Clock
}

func NewService(uow ports.UnitOfWork, clock ports.Clock) *Service {
	return &Service{uow: uow, clock: clock}
}

func (s *Service) CreateRefund(req domain.RefundRequest, idempotencyKey, fingerprint string) (domain.RefundResponse, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return domain.RefundResponse{}, domain.NewError(400, "missing_idempotency_key", "idempotency key is required")
	}
	if strings.TrimSpace(req.ChargeID) == "" {
		return domain.RefundResponse{}, domain.NewError(400, "invalid_charge_id", "charge_id is required")
	}

	var response domain.RefundResponse
	err := s.uow.Do(func(tx ports.Transaction) error {
		key := "create_refund:" + idempotencyKey
		result, err := tx.WithIdempotency(key, fingerprint, func() (any, error) {
			charge, err := tx.GetCharge(req.ChargeID)
			if err != nil {
				return nil, err
			}
			if charge.CapturedAmount == 0 {
				return nil, domain.NewError(409, "invalid_charge_state", "charge must be captured before refunding")
			}

			remaining := charge.CapturedAmount - charge.RefundedAmount
			amount := domain.Amount(req.Amount)
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
			refund := domain.Refund{ID: tx.NextID("re"), ChargeID: charge.ID, PaymentIntentID: charge.PaymentIntentID, Amount: amount, Status: domain.RefundSucceeded, CreatedAt: now, UpdatedAt: now}
			charge.RefundedAmount += amount
			if charge.RefundedAmount == charge.CapturedAmount {
				charge.Status = domain.ChargeRefunded
			} else {
				charge.Status = domain.ChargePartiallyRefunded
			}
			charge.UpdatedAt = now
			tx.SaveRefund(refund)
			tx.SaveCharge(charge)
			_ = tx.Publish(domain.RefundCreatedEvent{Refund: refund, Charge: charge})
			return domain.RefundResponse{Refund: refund, Charge: charge}, nil
		})
		if err != nil {
			return err
		}
		response = result.(domain.RefundResponse)
		return nil
	})
	if err != nil {
		return domain.RefundResponse{}, err
	}
	return response, nil
}
