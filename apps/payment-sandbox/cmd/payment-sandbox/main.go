package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"payment-sandbox/internal/adapters/observability/logging"
	newrelictracing "payment-sandbox/internal/adapters/observability/tracing/newrelic"
	oteltracing "payment-sandbox/internal/adapters/observability/tracing/otel"
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

	otelShutdown, err := oteltracing.Setup(context.Background(), "payment-sandbox")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = otelShutdown(context.Background()) }()

	nrApp, err := newrelictracing.Setup("payment-sandbox", os.Getenv("NEW_RELIC_LICENSE_KEY"))
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
		svc = sandbox.NewPostgresService(db)
	} else {
		svc = sandbox.NewService()
	}
	handler := server.New(svc, server.WithLogger(logger), server.WithNewRelic(nrApp))

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
