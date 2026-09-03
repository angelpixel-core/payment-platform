package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var landing = template.Must(template.New("landing").Parse(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Payment Sandbox Docs</title>
    <style>
      body { font-family: system-ui, sans-serif; margin: 40px; line-height: 1.5; }
      a { display: block; margin: 0.35rem 0; }
      code { background: #f4f4f5; padding: 0.1rem 0.3rem; border-radius: 4px; }
    </style>
  </head>
  <body>
    <h1>Payment Sandbox Docs</h1>
    <p>OpenAPI and story docs served by a small Go-only static server.</p>
    <h2>OpenAPI</h2>
    <a href="/openapi/swagger-ui.html">Swagger UI</a>
    <a href="/openapi/redoc.html">Redoc</a>
    <a href="/openapi/payment-sandbox.v1.yaml">payment-sandbox.v1.yaml</a>
    <a href="/openapi/payment-sandbox.v1.json">payment-sandbox.v1.json</a>
    <h2>Stories</h2>
    <a href="/epics/001-payment-sandbox-integration-mvp/README.md">Epic README</a>
    <a href="/epics/001-payment-sandbox-integration-mvp/stories/03-webhook-delivery/README.md">03 Webhook Delivery</a>
    <a href="/epics/001-payment-sandbox-integration-mvp/stories/05-workflow-consumer-integration/README.md">05 Workflow Consumer Integration</a>
  </body>
</html>`))

func main() {
	addr := flag.String("addr", ":8081", "listen address")
	root := flag.String("root", "", "docs root directory (defaults to the nearest parent docs directory)")
	flag.Parse()

	docsRoot, err := resolveDocsRoot(*root)
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir(docsRoot))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			fs.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := landing.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	server := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("serving docs from %s on %s", docsRoot, *addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func resolveDocsRoot(explicit string) (string, error) {
	if explicit != "" {
		return filepath.Abs(explicit)
	}
	start, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(start, "docs")
		if info, err := os.Stat(filepath.Join(candidate, "openapi", "payment-sandbox.v1.yaml")); err == nil && !info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(start)
		if parent == start {
			break
		}
		start = parent
	}
	return "", &os.PathError{Op: "stat", Path: filepath.Join("docs", "openapi", "payment-sandbox.v1.yaml"), Err: os.ErrNotExist}
}
