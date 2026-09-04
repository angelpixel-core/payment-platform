package cli

import "testing"

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
