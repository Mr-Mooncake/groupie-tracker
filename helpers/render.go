package helpers

// Shared page rendering utilities to show HTML templates and errors.

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
)

// TemplatesDir is the folder with the HTML templates (tests change it).
var TemplatesDir = filepath.Join("web", "templates")

// RenderError shows error.html with the given status code and message.
func RenderError(w http.ResponseWriter, statusCode int, message string) {
	tmpl, err := template.ParseFiles(filepath.Join(TemplatesDir, "error.html"))
	if err != nil {
		http.Error(w, message, statusCode)
		return
	}

	var page bytes.Buffer // build the page in memory first
	data := ErrorData{StatusCode: statusCode, Message: message}
	if err := tmpl.Execute(&page, data); err != nil {
		http.Error(w, message, statusCode)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	page.WriteTo(w)
}

// RenderPage fills a template with data and sends it. The page is built in
// memory first, so a template error never sends half a page.
func RenderPage(w http.ResponseWriter, name string, data any) {
	tmpl, err := template.ParseFiles(filepath.Join(TemplatesDir, name))
	if err != nil {
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	var page bytes.Buffer
	if err := tmpl.Execute(&page, data); err != nil {
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page.WriteTo(w)
}
