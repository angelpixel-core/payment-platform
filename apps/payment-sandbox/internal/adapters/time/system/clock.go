package system

import (
	"time"

	"payment-sandbox/internal/ports"
)

type Clock struct{}

func NewClock() ports.Clock { return Clock{} }

func (Clock) Now() time.Time { return time.Now() }
