package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"payment-sandbox/internal/sandbox"
	"payment-sandbox/internal/server"
)

func main() {
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8080"
	}

	svc := sandbox.NewService()
	handler := server.New(svc)

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
