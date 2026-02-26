package utils

import "html/template"

func BaseTemplate() *template.Template {
	return template.Must(template.New("").Funcs(template.FuncMap{
		"formatFileSizePretty": FormatSizePretty,
		"isImage":              IsImage,
		"isText":               IsText,
		"safeHTML":             SafeHTML,
	}).ParseGlob("web/templates/**/*.html"))
}

// SafeHTML marks a string as trusted HTML for use in templates.
func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}
