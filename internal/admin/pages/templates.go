package pages

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"sync"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

// sharedFuncMap contains template functions available to all admin pages.
var sharedFuncMap = template.FuncMap{
	"statusRu": StatusRu,
	"jsEscape": JsEscape,
	"stars":    Stars,
}

// Stars returns a visual star rating string (filled and empty stars).
func Stars(n int) string {
	if n < 0 {
		n = 0
	} else if n > 5 {
		n = 5
	}
	return strings.Repeat("★", n) + strings.Repeat("☆", 5-n)
}

// StatusRu maps English status strings to Russian translations.
func StatusRu(status string) string {
	switch status {
	case "pending":
		return "Ожидает"
	case "confirmed":
		return "Подтверждено"
	case "approved":
		return "Одобрено"
	case "rejected":
		return "Отклонено"
	case "cancelled":
		return "Отменено"
	case "completed":
		return "Завершено"
	case "hidden":
		return "Скрыто"
	default:
		return status
	}
}

// JsEscape escapes a string for safe embedding in JavaScript string literals.
// Returns template.JS to prevent html/template from applying additional
// context-aware escaping on top of our escaping.
func JsEscape(s string) template.JS {
	r := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		`'`, `\'`,
		`<`, `\x3c`,
		`>`, `\x3e`,
		`&`, `\x26`,
		"\n", `\n`,
		"\r", `\r`,
	)
	return template.JS(r.Replace(s))
}

// ParsePageTemplate creates a template from the shared base layout, components,
// and a page-specific template file.
func ParsePageTemplate(funcMap template.FuncMap, pageFile string) *template.Template {
	merged := template.FuncMap{}
	for k, v := range sharedFuncMap {
		merged[k] = v
	}
	for k, v := range funcMap {
		merged[k] = v
	}
	return template.Must(
		template.New("").Funcs(merged).ParseFS(templatesFS,
			"templates/base.tmpl",
			"templates/components.tmpl",
			pageFile,
		),
	)
}

var (
	staticHandler     http.Handler
	staticHandlerOnce sync.Once
)

// ServeStatic serves files from the embedded static directory.
func ServeStatic(w http.ResponseWriter, r *http.Request) {
	staticHandlerOnce.Do(func() {
		sub, _ := fs.Sub(staticFS, "static")
		staticHandler = http.StripPrefix("/static/", http.FileServer(http.FS(sub)))
	})
	staticHandler.ServeHTTP(w, r)
}
