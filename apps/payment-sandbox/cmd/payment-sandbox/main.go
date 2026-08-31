package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"payment-sandbox/internal/adapters/observability/logging"
	"payment-sandbox/internal/adapters/observability/metrics"
	newrelicmetrics "payment-sandbox/internal/adapters/observability/metrics/newrelic"
	otelmetrics "payment-sandbox/internal/adapters/observability/metrics/otel"
	"payment-sandbox/internal/adapters/persistence/postgres"
	"payment-sandbox/internal/sandbox"
	"payment-sandbox/internal/server"
)

func main() {
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8080"
	}

	logger := logging.NewJSON(os.Stdout, slog.LevelInfo)

	otelShutdown, err := otelmetrics.Setup(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = otelShutdown(context.Background()) }()

	nrApp, err := newrelicmetrics.Setup("payment-sandbox", os.Getenv("NEW_RELIC_LICENSE_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	recorder, err := metrics.NewRecorder(nrApp)
	if err != nil {
		log.Fatal(err)
	}

	var svc *sandbox.Service
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err := postgres.Open(context.Background(), dsn)
		if err != nil {
			log.Fatal(err)
		}
		if err := postgres.EnsureSchema(context.Background(), db); err != nil {
			log.Fatal(err)
		}
		svc = sandbox.NewPostgresServiceWithMetrics(db, recorder)
	} else {
		svc = sandbox.NewServiceWithMetrics(recorder)
	}
	handler := server.New(svc, server.WithLogger(logger), server.WithNewRelic(nrApp), server.WithMetrics(recorder))

	server := &http.Server{
		Addr:              ":" + addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("payment sandbox listening on :%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
