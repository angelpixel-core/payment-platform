package main

import (
	"context"
	"os"

	"payment-sandbox/internal/adapters/messaging/inprocess"
	"payment-sandbox/internal/adapters/persistence/memory"
	"payment-sandbox/internal/adapters/persistence/postgres"
	"payment-sandbox/internal/application/operations"
	"payment-sandbox/internal/cli"
)

func main() {
	runner := operations.ScenarioRunner(operations.NoopRunner{})
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err := postgres.Open(context.Background(), dsn)
		if err != nil {
			os.Exit(1)
		}
		defer db.Close()
		if err := postgres.EnsureSchema(context.Background(), db); err != nil {
			os.Exit(1)
		}
		runner = operations.NewSeedRunner(postgres.NewUnitOfWork(db, inprocess.NewPublisher(), nil))
	} else {
		runner = operations.NewSeedRunner(memory.NewUnitOfWork(memory.NewStore(nil), inprocess.NewPublisher()))
	}
	if err := cli.NewRootCmdWithRunner(runner).Execute(); err != nil {
		os.Exit(1)
	}
}
