package payments

import (
	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type PaymentQueryService struct {
	store ports.Store
}

func NewPaymentQueryService(store ports.Store) *PaymentQueryService {
	return &PaymentQueryService{store: store}
}

func (s *PaymentQueryService) GetPaymentIntent(id string) (PaymentIntentView, error) {
	intent, err := s.store.GetPaymentIntent(id)
	if err != nil {
		return PaymentIntentView{}, err
	}
	return PaymentIntentView{
		ID:              intent.ID,
		Status:          string(intent.Status),
		Amount:          intent.Amount.Int64(),
		Currency:        intent.Currency.String(),
		ChargeID:        intent.ChargeID,
		LatestAttemptID: intent.LatestAttemptID,
		CreatedAt:       intent.CreatedAt,
		UpdatedAt:       intent.UpdatedAt,
	}, nil
}

func (s *PaymentQueryService) GetPaymentAttempt(id string) (PaymentAttemptView, error) {
	attempt, err := s.store.GetPaymentAttempt(id)
	if err != nil {
		return PaymentAttemptView{}, err
	}
	return PaymentAttemptView{
		ID:                 attempt.ID,
		PaymentIntentID:    attempt.PaymentIntentID,
		PaymentMethodToken: attempt.PaymentMethodToken,
		Status:             string(attempt.Status),
		DeclineCode:        attempt.DeclineCode,
		ProcessorReference: attempt.ProcessorReference,
		RequestedAt:        attempt.RequestedAt,
		RespondedAt:        attempt.RespondedAt,
	}, nil
}

func (s *PaymentQueryService) GetCharge(id string) (ChargeView, error) {
	charge, err := s.store.GetCharge(id)
	if err != nil {
		return ChargeView{}, err
	}
	return ChargeView{
		ID:               charge.ID,
		PaymentIntentID:  charge.PaymentIntentID,
		PaymentAttemptID: charge.PaymentAttemptID,
		Amount:           charge.Amount.Int64(),
		CapturedAmount:   charge.CapturedAmount.Int64(),
		RefundedAmount:   charge.RefundedAmount.Int64(),
		Status:           string(charge.Status),
		CreatedAt:        charge.CreatedAt,
		UpdatedAt:        charge.UpdatedAt,
	}, nil
}

func (s *PaymentQueryService) GetRefund(id string) (RefundView, error) {
	refund, err := s.store.GetRefund(id)
	if err != nil {
		return RefundView{}, err
	}
	return RefundView{
		ID:              refund.ID,
		ChargeID:        refund.ChargeID,
		PaymentIntentID: refund.PaymentIntentID,
		Amount:          refund.Amount.Int64(),
		Status:          string(refund.Status),
		CreatedAt:       refund.CreatedAt,
		UpdatedAt:       refund.UpdatedAt,
	}, nil
}

func (s *PaymentQueryService) GetPaymentLifecycle(id string) (PaymentLifecycleView, error) {
	intent, err := s.store.GetPaymentIntent(id)
	if err != nil {
		return PaymentLifecycleView{}, err
	}

	view := PaymentLifecycleView{
		PaymentIntent: PaymentIntentView{
			ID:              intent.ID,
			Status:          string(intent.Status),
			Amount:          intent.Amount.Int64(),
			Currency:        intent.Currency.String(),
			ChargeID:        intent.ChargeID,
			LatestAttemptID: intent.LatestAttemptID,
			CreatedAt:       intent.CreatedAt,
			UpdatedAt:       intent.UpdatedAt,
		},
	}

	if intent.LatestAttemptID != "" {
		attempt, err := s.store.GetPaymentAttempt(intent.LatestAttemptID)
		if err != nil {
			return PaymentLifecycleView{}, err
		}
		attemptView := PaymentAttemptView{
			ID:                 attempt.ID,
			PaymentIntentID:    attempt.PaymentIntentID,
			PaymentMethodToken: attempt.PaymentMethodToken,
			Status:             string(attempt.Status),
			DeclineCode:        attempt.DeclineCode,
			ProcessorReference: attempt.ProcessorReference,
			RequestedAt:        attempt.RequestedAt,
			RespondedAt:        attempt.RespondedAt,
		}
		view.LatestAttempt = &attemptView
	}

	if intent.ChargeID != "" {
		charge, err := s.store.GetCharge(intent.ChargeID)
		if err != nil {
			return PaymentLifecycleView{}, err
		}
		chargeView := ChargeView{
			ID:               charge.ID,
			PaymentIntentID:  charge.PaymentIntentID,
			PaymentAttemptID: charge.PaymentAttemptID,
			Amount:           charge.Amount.Int64(),
			CapturedAmount:   charge.CapturedAmount.Int64(),
			RefundedAmount:   charge.RefundedAmount.Int64(),
			Status:           string(charge.Status),
			CreatedAt:        charge.CreatedAt,
			UpdatedAt:        charge.UpdatedAt,
		}
		view.Charge = &chargeView
		remaining := charge.CapturedAmount - charge.RefundedAmount
		if remaining > 0 {
			view.IsRefundable = true
			view.RefundableAmount = remaining.Int64()
		}
	}

	view.Status = derivePaymentLifecycleStatus(intent.Status, view.Charge)
	return view, nil
}

func derivePaymentLifecycleStatus(intentStatus domain.PaymentIntentStatus, charge *ChargeView) string {
	switch intentStatus {
	case domain.PaymentIntentRequiresPaymentMethod, domain.PaymentIntentRequiresConfirmation, domain.PaymentIntentRequiresAction, domain.PaymentIntentProcessing, domain.PaymentIntentRequiresCapture, domain.PaymentIntentFailed:
		return string(intentStatus)
	case domain.PaymentIntentSucceeded:
		if charge == nil {
			return string(intentStatus)
		}
		switch charge.Status {
		case string(domain.ChargeRefunded):
			return "refunded"
		case string(domain.ChargePartiallyRefunded):
			return "partially_refunded"
		default:
			return string(intentStatus)
		}
	default:
		return string(intentStatus)
	}
}
