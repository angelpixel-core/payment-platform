package commands_test

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

func TestCommandImportBoundaries(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller path")
	}

	root := filepath.Dir(currentFile)
	allowed := map[string]struct{}{
		"payment-sandbox/internal/domain":             {},
		"payment-sandbox/internal/ports":              {},
		"payment-sandbox/internal/application/support": {},
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

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
			if strings.HasPrefix(importPath, "payment-sandbox/internal/") {
				if _, ok := allowed[importPath]; !ok {
					return &boundaryError{file: path, importPath: importPath}
				}
			}
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

type boundaryError struct {
	file       string
	importPath string
}

func (e *boundaryError) Error() string {
	return "forbidden internal import " + e.importPath + " in " + e.file
}
