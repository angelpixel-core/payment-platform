package memory

import (
	"testing"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

func BenchmarkMemoryStoreSaveGetPaymentIntent(b *testing.B) {
	store := NewStore(nil)
	intent := domain.PaymentIntent{ID: "pi_bench", Currency: "USD"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.SavePaymentIntent(intent)
		if _, err := store.GetPaymentIntent(intent.ID); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemoryUnitOfWorkDo(b *testing.B) {
	store := NewStore(nil)
	uow := NewUnitOfWork(store, benchmarkPublisher{})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := uow.Do(func(tx ports.Transaction) error {
			tx.SavePaymentIntent(domain.PaymentIntent{ID: "pi_bench"})
			return nil
		}); err != nil {
			b.Fatal(err)
		}
	}
}

type benchmarkPublisher struct{}

func (benchmarkPublisher) Publish(domain.Event) error           { return nil }
func (benchmarkPublisher) Subscribe(string, ports.EventHandler) {}
