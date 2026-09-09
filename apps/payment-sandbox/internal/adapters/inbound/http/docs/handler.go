package docs

import (
	"net/http"
	"os"
	"path/filepath"
)

const defaultRoot = "docs/openapi"

func New(root string) http.Handler {
	if root == "" {
		root = defaultRoot
	}
	root, _ = filepath.Abs(root)
	files := http.FileServer(http.Dir(root))
	return http.StripPrefix("/openapi/", files)
}

func Exists(root string) bool {
	if root == "" {
		root = defaultRoot
	}
	info, err := os.Stat(root)
	return err == nil && info.IsDir()
}
