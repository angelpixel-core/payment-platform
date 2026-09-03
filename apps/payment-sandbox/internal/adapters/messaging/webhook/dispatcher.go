package webhook

import (
	"context"
	"errors"
	"time"
)

var ErrNoTransport = errors.New("webhook transport is required")

const defaultMaxAttempts = 3

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
	transport   Transport
	maxAttempts int
	backoff     func(attempt int) time.Duration
	sleep       func(time.Duration)
}

func NewDispatcher(transport Transport) *Dispatcher {
	return &Dispatcher{
		transport:   transport,
		maxAttempts: defaultMaxAttempts,
		backoff:     simpleBackoff,
		sleep:       time.Sleep,
	}
}

func simpleBackoff(attempt int) time.Duration {
	return time.Duration(attempt) * 10 * time.Millisecond
}

func (d *Dispatcher) Dispatch(ctx context.Context, delivery Delivery) error {
	if d == nil || d.transport == nil {
		return ErrNoTransport
	}
	if d.maxAttempts < 1 {
		d.maxAttempts = 1
	}
	if d.backoff == nil {
		d.backoff = simpleBackoff
	}
	if d.sleep == nil {
		d.sleep = func(time.Duration) {}
	}

	var lastErr error
	request := Request{Endpoint: delivery.Endpoint, Body: delivery.Payload}
	for attempt := 1; attempt <= d.maxAttempts; attempt++ {
		if err := d.transport.Send(ctx, request); err != nil {
			lastErr = err
			if attempt == d.maxAttempts {
				break
			}
			d.sleep(d.backoff(attempt))
			continue
		}
		return nil
	}
	return lastErr
}
