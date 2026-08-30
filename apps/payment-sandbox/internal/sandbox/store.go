package sandbox

import "payment-sandbox/internal/adapters/memory"

type Store = memory.MemoryStore
type MemoryStore = memory.MemoryStore

func NewMemoryStore() *MemoryStore { return memory.NewStore() }
