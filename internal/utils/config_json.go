package utils

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
)

const Null = "null"

// ConfigJSON returns compact JSON for embedding in pages (copy / prefill).
func ConfigJSON(v any) template.JS {
	normalized := normalizeJSONValue(v)
	if normalized == nil {
		return Null
	}
	b, err := json.Marshal(normalized)
	if err != nil {
		return Null
	}
	safe := strings.ReplaceAll(string(b), "<", `\u003c`)
	return template.JS(safe)
}

// FormatConfigPretty returns indented JSON for display in a code block.
func FormatConfigPretty(v any) string {
	normalized := normalizeJSONValue(v)
	if normalized == nil {
		return ""
	}
	b, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}

// HasConfiguration reports whether a stored config value has usable content.
func HasConfiguration(v any) bool {
	return normalizeJSONValue(v) != nil
}

func normalizeJSONValue(v any) any {
	if v == nil {
		return nil
	}

	switch typed := v.(type) {
	case string:
		if typed == "" {
			return nil
		}
		var parsed any
		if err := json.Unmarshal([]byte(typed), &parsed); err == nil {
			return parsed
		}
		return typed
	case []byte:
		if len(typed) == 0 {
			return nil
		}
		var parsed any
		if err := json.Unmarshal(typed, &parsed); err == nil {
			return parsed
		}
		return string(typed)
	case map[string]any:
		return typed
	case map[any]any:
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			out[fmt.Sprint(key)] = value
		}
		return out
	default:
		b, err := json.Marshal(typed)
		if err != nil {
			return typed
		}
		var parsed any
		if err := json.Unmarshal(b, &parsed); err != nil {
			return typed
		}
		return parsed
	}
}
