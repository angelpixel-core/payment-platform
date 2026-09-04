package cli

import (
	"context"
	"errors"
	"testing"

	"payment-sandbox/internal/application/operations"
)

type fakeRunner struct {
	simulateScenario string
	replayScenario   string
	burstCalled      bool
	seedCalled       bool
}

func (f *fakeRunner) Simulate(ctx context.Context, scenario string) error {
	f.simulateScenario = scenario
	return nil
}

func (f *fakeRunner) Replay(ctx context.Context, scenario string) error {
	f.replayScenario = scenario
	return nil
}
func (f *fakeRunner) Burst(ctx context.Context) error {
	f.burstCalled = true
	return nil
}
func (f *fakeRunner) Seed(ctx context.Context) error {
	f.seedCalled = true
	return nil
}

func TestNewRootCmdRegistersOperationalSubcommands(t *testing.T) {
	cmd := NewRootCmd()
	if cmd.Use != "payment-sandbox" {
		t.Fatalf("unexpected root use: %s", cmd.Use)
	}

	want := map[string]bool{
		"serve":    false,
		"simulate": false,
		"replay":   false,
		"burst":    false,
		"seed":     false,
	}
	for _, sub := range cmd.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("expected subcommand %q to be registered", name)
		}
	}
}

func TestScenarioCommandsCallRunner(t *testing.T) {
	runner := &fakeRunner{}
	cmd := NewRootCmdWithRunner(runner)

	if err := executeCommand(cmd, []string{"simulate", "approved_immediate"}); err != nil {
		t.Fatalf("simulate command failed: %v", err)
	}
	if runner.simulateScenario != "approved_immediate" {
		t.Fatalf("simulate scenario not forwarded: %q", runner.simulateScenario)
	}

	runner = &fakeRunner{}
	cmd = NewRootCmdWithRunner(runner)
	if err := executeCommand(cmd, []string{"replay", "retry_webhook"}); err != nil {
		t.Fatalf("replay command failed: %v", err)
	}
	if runner.replayScenario != "retry_webhook" {
		t.Fatalf("replay scenario not forwarded: %q", runner.replayScenario)
	}

	runner = &fakeRunner{}
	cmd = NewRootCmdWithRunner(runner)
	if err := executeCommand(cmd, []string{"burst"}); err != nil {
		t.Fatalf("burst command failed: %v", err)
	}
	if !runner.burstCalled {
		t.Fatal("burst runner was not called")
	}

	runner = &fakeRunner{}
	cmd = NewRootCmdWithRunner(runner)
	if err := executeCommand(cmd, []string{"seed"}); err != nil {
		t.Fatalf("seed command failed: %v", err)
	}
	if !runner.seedCalled {
		t.Fatal("seed runner was not called")
	}
}

func TestScenarioCommandsUseNoopRunnerByDefault(t *testing.T) {
	cmd := NewRootCmdWithRunner(nil)
	if err := executeCommand(cmd, []string{"simulate", "approved_immediate"}); !errors.Is(err, operations.ErrNotImplemented) {
		t.Fatalf("expected noop runner error, got %v", err)
	}
}

func executeCommand(cmd interface{ SetArgs([]string); Execute() error }, args []string) error {
	cmd.SetArgs(args)
	return cmd.Execute()
}
