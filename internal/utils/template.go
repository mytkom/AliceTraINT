package utils

import "html/template"

func BaseTemplate() *template.Template {
	return template.Must(template.New("").Funcs(template.FuncMap{
		"formatFileSizePretty": FormatSizePretty,
	}).ParseGlob("web/templates/**/*.html"))
}
