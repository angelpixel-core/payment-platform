package outbox

import (
	"testing"

	"payment-sandbox/internal/adapters/messaging/inprocess"
	"payment-sandbox/internal/domain"
)

type benchEvent struct{ name string }

func (e benchEvent) EventName() string { return e.name }
func (e benchEvent) EventVersion() int  { return 1 }

func BenchmarkOutboxPublish(b *testing.B) {
	downstream := inprocess.NewPublisher()
	downstream.Subscribe("bench.event", func(domain.Event) error { return nil })
	publisher := NewPublisher(downstream, nil)
	event := benchEvent{name: "bench.event"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := publisher.Publish(event); err != nil {
			b.Fatal(err)
		}
	}
}
