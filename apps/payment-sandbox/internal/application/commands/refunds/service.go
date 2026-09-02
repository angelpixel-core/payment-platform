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
			now := s.clock.Now()
			refund, err := charge.Refund(domain.RefundChargeCommand{RefundID: tx.NextID("re"), Amount: domain.Amount(req.Amount), Now: now})
			if err != nil {
				return nil, err
			}
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
