package webhook

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrNoTransport = errors.New("webhook transport is required")
var ErrDeliveryFailed = errors.New("webhook delivery failed after retries")

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

type DispatchError struct {
	DeliveryID string
	Attempts   int
	Err        error
}

func (e DispatchError) Error() string {
	return fmt.Sprintf("%s: delivery_id=%s attempts=%d: %v", ErrDeliveryFailed.Error(), e.DeliveryID, e.Attempts, e.Err)
}

func (e DispatchError) Unwrap() error { return e.Err }

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
	return DispatchError{DeliveryID: delivery.DeliveryID, Attempts: d.maxAttempts, Err: lastErr}
}
