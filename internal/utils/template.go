package utils

import "html/template"

func BaseTemplate() *template.Template {
	return template.Must(template.New("").Funcs(template.FuncMap{
		"formatFileSizePretty": FormatSizePretty,
		"isImage":              IsImage,
		"isText":               IsText,
	}).ParseGlob("web/templates/**/*.html"))
}
