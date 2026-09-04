package main

import (
	"os"

	"payment-sandbox/internal/application/operations"
	"payment-sandbox/internal/cli"
)

func main() {
	if err := cli.NewRootCmdWithRunner(operations.NoopRunner{}).Execute(); err != nil {
		os.Exit(1)
	}
}
