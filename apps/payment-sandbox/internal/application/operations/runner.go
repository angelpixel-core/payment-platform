package operations

import (
	"context"
	"errors"
)

var ErrNotImplemented = errors.New("operations runner is not implemented yet")

type ScenarioRunner interface {
	Simulate(ctx context.Context, scenario string) error
	Replay(ctx context.Context, scenario string) error
	Burst(ctx context.Context) error
	Seed(ctx context.Context) error
}

type NoopRunner struct{}

func (NoopRunner) Simulate(ctx context.Context, scenario string) error { return ErrNotImplemented }
func (NoopRunner) Replay(ctx context.Context, scenario string) error   { return ErrNotImplemented }
func (NoopRunner) Burst(ctx context.Context) error                     { return ErrNotImplemented }
func (NoopRunner) Seed(ctx context.Context) error                      { return ErrNotImplemented }
