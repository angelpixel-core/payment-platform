package server

import (
	"net/http"

	httpadapter "payment-sandbox/internal/adapters/inbound/http"
	"payment-sandbox/internal/sandbox"
)

func New(svc *sandbox.Service) http.Handler {
	return httpadapter.New(svc)
}
