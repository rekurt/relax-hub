package pages

import (
	"embed"
	"html/template"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// ParsePageTemplate creates a template from the shared base layout, components,
// and a page-specific template file.
func ParsePageTemplate(funcMap template.FuncMap, pageFile string) *template.Template {
	return template.Must(
		template.New("").Funcs(funcMap).ParseFS(templatesFS,
			"templates/base.tmpl",
			"templates/components.tmpl",
			pageFile,
		),
	)
}
