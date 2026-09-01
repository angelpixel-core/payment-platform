package sandbox

import (
	"payment-sandbox/internal/adapters/observability/metrics"
	"payment-sandbox/internal/adapters/persistence/memory"
)

type Store = memory.MemoryStore
type MemoryStore = memory.MemoryStore

func NewMemoryStore(recorder metrics.MetricsRecorder) *MemoryStore { return memory.NewStore(recorder) }
