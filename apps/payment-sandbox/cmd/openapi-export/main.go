package main

import (
	"flag"
	"fmt"
	"os"

	openapi "payment-sandbox/internal/docs/openapi"
)

func main() {
	in := flag.String("in", "../../docs/openapi/payment-sandbox.v1.yaml", "input OpenAPI YAML path")
	out := flag.String("out", "../../docs/openapi/payment-sandbox.v1.json", "output OpenAPI JSON path")
	flag.Parse()

	jsonBytes, err := openapi.GenerateJSONFile(*in)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, jsonBytes, 0o644); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, fmt.Errorf("write openapi json: %w", err))
		os.Exit(1)
	}
}
