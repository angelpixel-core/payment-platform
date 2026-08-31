package server

import (
	"log/slog"

	nr "github.com/newrelic/go-agent/v3/newrelic"
	"net/http"

	httpadapter "payment-sandbox/internal/adapters/inbound/http"
	"payment-sandbox/internal/sandbox"
)

type config struct {
	logger *slog.Logger
	nrApp  *nr.Application
}

type Option func(*config)

func WithLogger(logger *slog.Logger) Option {
	return func(cfg *config) {
		if logger != nil {
			cfg.logger = logger
		}
	}
}

func WithNewRelic(app *nr.Application) Option {
	return func(cfg *config) {
		cfg.nrApp = app
	}
}

func New(svc *sandbox.Service, opts ...Option) http.Handler {
	cfg := config{logger: slog.Default()}
	for _, opt := range opts {
		opt(&cfg)
	}
	h := httpadapter.New(svc)
	h = httpadapter.Observability(h, cfg.logger)
	if cfg.nrApp != nil {
		_, h = nr.WrapHandle(cfg.nrApp, "payment-sandbox", h)
	}
	return h
}
