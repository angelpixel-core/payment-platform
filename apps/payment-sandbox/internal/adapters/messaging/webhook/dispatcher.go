package webhook

import (
	"context"
	"errors"
)

var ErrNoTransport = errors.New("webhook transport is required")

type Transport interface {
	Send(ctx context.Context, request Request) error
}

type Request struct {
	Endpoint string
	Body     []byte
}

type Delivery struct {
	EventType  string
	EventID    string
	DeliveryID string
	Attempt    int
	Endpoint   string
	Payload    []byte
}

func NewDelivery(eventType, eventID, deliveryID, endpoint string, attempt int, payload []byte) Delivery {
	return Delivery{
		EventType:  eventType,
		EventID:    eventID,
		DeliveryID: deliveryID,
		Attempt:    attempt,
		Endpoint:   endpoint,
		Payload:    append([]byte(nil), payload...),
	}
}

// Dispatcher is the webhook delivery entrypoint.
// It is intentionally small here; payload construction and retries are added later.
type Dispatcher struct {
	transport Transport
}

func NewDispatcher(transport Transport) *Dispatcher {
	return &Dispatcher{transport: transport}
}

func (d *Dispatcher) Dispatch(ctx context.Context, delivery Delivery) error {
	if d == nil || d.transport == nil {
		return ErrNoTransport
	}
	return d.transport.Send(ctx, Request{Endpoint: delivery.Endpoint, Body: delivery.Payload})
}
