package openapi

import (
	"encoding/json"
	"fmt"
	"os"

	yaml "go.yaml.in/yaml/v3"
)

// GenerateJSON converts an OpenAPI YAML document into canonical JSON bytes.
func GenerateJSON(yamlBytes []byte) ([]byte, error) {
	var doc any
	if err := yaml.Unmarshal(yamlBytes, &doc); err != nil {
		return nil, fmt.Errorf("parse openapi yaml: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(normalizeYAMLValue(doc), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode openapi json: %w", err)
	}

	return append(jsonBytes, '\n'), nil
}

// GenerateJSONFile reads an OpenAPI YAML file and returns the generated JSON.
func GenerateJSONFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read openapi yaml: %w", err)
	}
	return GenerateJSON(content)
}

func normalizeYAMLValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			out[key] = normalizeYAMLValue(item)
		}
		return out
	case map[any]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			out[fmt.Sprint(key)] = normalizeYAMLValue(item)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = normalizeYAMLValue(item)
		}
		return out
	default:
		return v
	}
}
