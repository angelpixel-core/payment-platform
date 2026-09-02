package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const internalImportPrefix = "payment-sandbox/internal/"

func TestCoreLayerImportBoundaries(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller path")
	}

	root := filepath.Dir(currentFile)
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		layer := topLevelLayer(path, root)
		if layer == "" {
			return nil
		}

		allowed := allowedInternalImports(layer)
		fset := token.NewFileSet()
		file, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}

		for _, spec := range file.Imports {
			importPath, unquoteErr := strconv.Unquote(spec.Path.Value)
			if unquoteErr != nil {
				return unquoteErr
			}
			if !strings.HasPrefix(importPath, internalImportPrefix) {
				continue
			}
			if !isAllowedImport(importPath, allowed) {
				return &boundaryError{file: path, layer: layer, importPath: importPath}
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func topLevelLayer(path, root string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return ""
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) == 0 {
		return ""
	}
	switch parts[0] {
	case "domain", "ports", "application", "adapters":
		return parts[0]
	case "server", "sandbox":
		return ""
	default:
		return ""
	}
}

func allowedInternalImports(layer string) []string {
	switch layer {
	case "domain":
		return nil
	case "ports":
		return []string{internalImportPrefix + "domain"}
	case "application":
		return []string{
			internalImportPrefix + "domain",
			internalImportPrefix + "ports",
			internalImportPrefix + "application",
		}
	case "adapters":
		return []string{
			internalImportPrefix + "domain",
			internalImportPrefix + "ports",
			internalImportPrefix + "adapters",
			internalImportPrefix + "sandbox",
		}
	default:
		return nil
	}
}

func isAllowedImport(importPath string, allowed []string) bool {
	for _, prefix := range allowed {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}
	return false
}

type boundaryError struct {
	file       string
	layer      string
	importPath string
}

func (e *boundaryError) Error() string {
	return "forbidden internal import " + e.importPath + " in " + e.file + " (layer " + e.layer + ")"
}
