package helpers

// [Hashim] Page handlers and error rendering.

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
)

var templatesDir = "templates"

type errorData struct {
	StatusCode int
	StatusText string
	Message    string
}

// renderError shows templates/error.html; falls back to plain text.
func renderError(w http.ResponseWriter, status int, message string) {
	tmpl, err := template.ParseFiles(filepath.Join(templatesDir, "error.html"))
	if err != nil {
		http.Error(w, message, status)
		return
	}
	var buf bytes.Buffer
	data := errorData{StatusCode: status, StatusText: http.StatusText(status), Message: message}
	if err := tmpl.Execute(&buf, data); err != nil {
		http.Error(w, message, status)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

// render executes a template into a buffer first, so a template failure
// never sends a half-written page.
func render(w http.ResponseWriter, name string, data interface{}) {
	funcs := template.FuncMap{"formatLocation": FormatLocation}
	tmpl, err := template.New(name).Funcs(funcs).ParseFiles(filepath.Join(templatesDir, name))
	if err != nil {
		renderError(w, http.StatusInternalServerError, "Failed to load the template")
		return
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		renderError(w, http.StatusInternalServerError, "Failed to render the page")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

// HandleHome shows all artists as cards.
func HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		renderError(w, http.StatusNotFound, "Page not found")
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		renderError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	artists, _, err := searchCache.get()
	if err != nil {
		renderError(w, http.StatusInternalServerError, "Could not load artists, please try again later")
		return
	}
	render(w, "index.html", artists)
}

// HandleArtist shows one artist with members and concert locations.
func HandleArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		renderError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		renderError(w, http.StatusBadRequest, "Invalid artist id")
		return
	}

	artists, relations, err := searchCache.get()
	if err != nil {
		renderError(w, http.StatusInternalServerError, "Could not load artist data, please try again later")
		return
	}

	detail := ArtistDetail{}
	found := false
	for _, a := range artists {
		if a.ID == id {
			detail.Artist = a
			found = true
			break
		}
	}
	if !found {
		renderError(w, http.StatusNotFound, "Artist not found")
		return
	}
	for _, rel := range relations {
		if rel.ID == id {
			detail.Relations = rel.DatesLocations
			break
		}
	}
	render(w, "artist.html", detail)
}
