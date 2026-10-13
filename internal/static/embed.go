package static

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var TemplateFS embed.FS

func SetUpFs() *template.Template {
	tmpl, err := template.ParseFS(TemplateFS, "templates/*.html")
	if err != nil {
		panic("Не удалось скомпилировать шаблоны: " + err.Error())
	}
	return tmpl
}
