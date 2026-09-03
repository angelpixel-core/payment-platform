package openapi

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGenerateJSONMatchesCheckedInFile(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	yamlPath := filepath.Join(root, "docs", "openapi", "payment-sandbox.v1.yaml")
	jsonPath := filepath.Join(root, "docs", "openapi", "payment-sandbox.v1.json")

	got, err := GenerateJSONFile(yamlPath)
	if err != nil {
		t.Fatalf("generate json failed: %v", err)
	}
	want, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read checked-in json failed: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("generated json is out of sync with %s", jsonPath)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", ".."))
}
