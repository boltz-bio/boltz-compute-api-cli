package structuredinclude

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
)

type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

func ParseReference(ref string) (Format, string, bool) {
	if path, ok := strings.CutPrefix(ref, "@json://"); ok {
		return FormatJSON, path, true
	}
	if path, ok := strings.CutPrefix(ref, "@yaml://"); ok {
		return FormatYAML, path, true
	}
	return "", "", false
}

func ParseBytes(format Format, path string, content []byte) (any, error) {
	var parsed any
	err := unmarshal(format, content, &parsed)
	if err != nil {
		return nil, fmt.Errorf("failed to parse @%s://%s: %w", format, path, err)
	}

	return parsed, nil
}

func ParseBytesAs[T any](format Format, path string, content []byte) (T, error) {
	var parsed T
	err := unmarshal(format, content, &parsed)
	if err != nil {
		return parsed, fmt.Errorf("failed to parse @%s://%s: %w", format, path, err)
	}
	return parsed, nil
}

func unmarshal(format Format, content []byte, dest any) error {
	switch format {
	case FormatJSON:
		return json.Unmarshal(content, dest)
	case FormatYAML:
		return yaml.Unmarshal(content, dest)
	default:
		return fmt.Errorf("unsupported structured include format %q", format)
	}
}
