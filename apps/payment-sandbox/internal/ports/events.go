package ports

import "payment-sandbox/internal/domain"

type EventHandler func(domain.Event) error

type EventPublisher interface {
	Publish(event domain.Event) error
	Subscribe(eventName string, handler EventHandler)
}
