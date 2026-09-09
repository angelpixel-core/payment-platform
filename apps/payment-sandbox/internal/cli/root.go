package cli

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"payment-sandbox/internal/adapters/observability/logging"
	"payment-sandbox/internal/adapters/observability/metrics"
	newrelicmetrics "payment-sandbox/internal/adapters/observability/metrics/newrelic"
	otelmetrics "payment-sandbox/internal/adapters/observability/metrics/otel"
	"payment-sandbox/internal/adapters/persistence/postgres"
	"payment-sandbox/internal/application/operations"
	"payment-sandbox/internal/bootstrap"
	"payment-sandbox/internal/sandbox"
)

const defaultPort = "8080"

type runtimeConfig struct {
	port               string
	databaseURL        string
	newRelicLicenseKey string
}

func NewRootCmd() *cobra.Command {
	return NewRootCmdWithRunner(operations.NoopRunner{})
}

func NewRootCmdWithRunner(runner operations.ScenarioRunner) *cobra.Command {
	if runner == nil {
		runner = operations.NoopRunner{}
	}
	cfg := runtimeConfig{
		port:               envOr("PORT", defaultPort),
		databaseURL:        os.Getenv("DATABASE_URL"),
		newRelicLicenseKey: os.Getenv("NEW_RELIC_LICENSE_KEY"),
	}

	cmd := &cobra.Command{
		Use:           "payment-sandbox",
		Short:         "Payment sandbox service and operational tooling",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd.Context(), cfg)
		},
	}

	cmd.PersistentFlags().StringVar(&cfg.port, "port", cfg.port, "HTTP port for the sandbox server")
	cmd.PersistentFlags().StringVar(&cfg.databaseURL, "database-url", cfg.databaseURL, "PostgreSQL DSN for the sandbox server")
	cmd.PersistentFlags().StringVar(&cfg.newRelicLicenseKey, "new-relic-license-key", cfg.newRelicLicenseKey, "New Relic license key for optional metrics sink")

	cmd.AddCommand(
		newServeCmd(cfg),
		newScenarioCommand("simulate <scenario>", "Run a deterministic sandbox scenario", cobra.ExactArgs(1), func(ctx context.Context, args []string) error {
			return runner.Simulate(ctx, args[0])
		}),
		newScenarioCommand("replay <scenario>", "Replay a sandbox scenario", cobra.ExactArgs(1), func(ctx context.Context, args []string) error {
			return runner.Replay(ctx, args[0])
		}),
		newScenarioCommand("burst", "Run a burst of sandbox requests", cobra.NoArgs, func(ctx context.Context, args []string) error {
			return runner.Burst(ctx)
		}),
		newScenarioCommand("seed", "Seed local sandbox data", cobra.NoArgs, func(ctx context.Context, args []string) error {
			return runner.Seed(ctx)
		}),
	)

	return cmd
}

func newServeCmd(cfg runtimeConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the sandbox HTTP API",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd.Context(), cfg)
		},
	}
}

func newScenarioCommand(use, short string, args cobra.PositionalArgs, run func(context.Context, []string) error) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  args,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args)
		},
	}
}

func runServe(ctx context.Context, cfg runtimeConfig) error {
	logger := logging.NewJSON(os.Stdout, slog.LevelInfo)

	otelShutdown, err := otelmetrics.Setup(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = otelShutdown(context.Background()) }()

	nrApp, err := newrelicmetrics.Setup("payment-sandbox", cfg.newRelicLicenseKey)
	if err != nil {
		return err
	}
	recorder, err := metrics.NewRecorder(nrApp)
	if err != nil {
		return err
	}

	var svc *sandbox.Service
	if cfg.databaseURL != "" {
		db, err := postgres.Open(ctx, cfg.databaseURL)
		if err != nil {
			return err
		}
		if err := postgres.EnsureSchema(ctx, db); err != nil {
			return err
		}
		svc = sandbox.NewPostgresServiceWithMetrics(db, recorder)
	} else {
		svc = sandbox.NewServiceWithMetrics(recorder)
	}

	handler := bootstrap.New(svc, bootstrap.WithLogger(logger), bootstrap.WithNewRelic(nrApp), bootstrap.WithMetrics(recorder), bootstrap.WithDocsRoot(os.Getenv("DOCS_ROOT")))

	server := &http.Server{
		Addr:              ":" + cfg.port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("payment sandbox listening on :%s", cfg.port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
