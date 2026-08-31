package sandbox

import "payment-sandbox/internal/adapters/persistence/memory"

type Store = memory.MemoryStore
type MemoryStore = memory.MemoryStore

func NewMemoryStore() *MemoryStore { return memory.NewStore() }
