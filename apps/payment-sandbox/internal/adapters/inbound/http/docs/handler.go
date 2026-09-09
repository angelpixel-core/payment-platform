package docs

import (
	"net/http"
	"os"
	"path/filepath"
)

const defaultRoot = "docs"

func New(root string) http.Handler {
	if root == "" {
		root = defaultRoot
	}
	root, _ = filepath.Abs(root)
	files := http.FileServer(http.Dir(root))
	openapi := http.FileServer(http.Dir(filepath.Join(root, "openapi")))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/docs" || r.URL.Path == "/docs/" {
			http.Redirect(w, r, "/docs/swagger/", http.StatusFound)
			return
		}
		if len(r.URL.Path) >= len("/openapi/") && r.URL.Path[:len("/openapi/")] == "/openapi/" {
			http.StripPrefix("/openapi/", openapi).ServeHTTP(w, r)
			return
		}
		http.StripPrefix("/docs/", files).ServeHTTP(w, r)
	})
}

func Exists(root string) bool {
	if root == "" {
		root = defaultRoot
	}
	info, err := os.Stat(filepath.Join(root, "openapi", "v1"))
	return err == nil && info.IsDir()
}
