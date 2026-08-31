package observability

import (
	"sync"

	"payment-sandbox/internal/domain"
	"payment-sandbox/internal/ports"
)

type Recorder struct {
	mu     sync.Mutex
	names  []string
	counts map[string]int
}

func NewRecorder() *Recorder {
	return &Recorder{counts: make(map[string]int)}
}

func (r *Recorder) Handle(event domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.names = append(r.names, event.EventName())
	r.counts[event.EventName()]++
	return nil
}

func (r *Recorder) Snapshot() (names []string, counts map[string]int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	names = append([]string(nil), r.names...)
	counts = make(map[string]int, len(r.counts))
	for k, v := range r.counts {
		counts[k] = v
	}
	return names, counts
}

func RegisterInternalHandlers(pub ports.EventPublisher, recorder *Recorder) {
	if pub == nil || recorder == nil {
		return
	}
	for _, eventName := range []string{
		"payment_intent.created",
		"payment_intent.confirmed",
		"payment_intent.finalized",
		"payment_intent.captured",
		"refund.created",
	} {
		pub.Subscribe(eventName, recorder.Handle)
	}
}
