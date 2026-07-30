package utils

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/mytkom/AliceTraINT/internal/jalien"
)

func BaseTemplate() *template.Template {
	return template.Must(template.New("").Funcs(template.FuncMap{
		"formatFileSizePretty":       FormatSizePretty,
		"isImage":                    IsImage,
		"isText":                     IsText,
		"safeHTML":                   SafeHTML,
		"summarizeAODByPeriod":       jalien.SummarizeByPeriod,
		"summarizeAODByPeriodAndRun": jalien.SummarizeByPeriodAndRun,
		"configFieldName":            ConfigFieldName,
		"configPath":                 ConfigPath,
		"formatListDefault":          FormatListDefault,
		"rangeDefaultAt":             RangeDefaultAt,
		"showWhenEqualsJSON":         ShowWhenEqualsJSON,
		"optionSelected":             OptionSelected,
		"configJSON":                 ConfigJSON,
		"formatConfigPretty":         FormatConfigPretty,
	}).ParseGlob("web/templates/**/*.html"))
}

func ConfigPath(path, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}

func ConfigFieldName(path, name string) string {
	return "configuration." + ConfigPath(path, name)
}

func FormatListDefault(v interface{}) string {
	if v == nil {
		return ""
	}
	switch vals := v.(type) {
	case []interface{}:
		parts := make([]string, 0, len(vals))
		for _, item := range vals {
			parts = append(parts, fmt.Sprint(item))
		}
		return strings.Join(parts, ", ")
	case []string:
		return strings.Join(vals, ", ")
	default:
		return fmt.Sprint(v)
	}
}

func RangeDefaultAt(v interface{}, index int) interface{} {
	switch vals := v.(type) {
	case []interface{}:
		if index >= 0 && index < len(vals) {
			return vals[index]
		}
	case []float64:
		if index >= 0 && index < len(vals) {
			return vals[index]
		}
	case []int:
		if index >= 0 && index < len(vals) {
			return vals[index]
		}
	}
	return ""
}

func ShowWhenEqualsJSON(v interface{}) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return template.JS(b)
}

func OptionSelected(optionValue, defaultValue interface{}) bool {
	if optionValue == nil && defaultValue == nil {
		return true
	}
	return fmt.Sprint(optionValue) == fmt.Sprint(defaultValue)
}

// SafeHTML marks a string as trusted HTML for use in templates.
func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}
