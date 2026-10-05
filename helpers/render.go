package helpers

// Shared page rendering, so both main.go and the helpers files can show pages and errors.

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
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

func HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		RenderError(w, http.StatusNotFound, "404 Page Not Found")
		return
	}

	if r.Method != http.MethodGet {
		RenderError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}

	artists, err := FetchAllArtists()
	if err != nil {
		log.Printf("Error fetching artists: %v", err)
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	RenderPage(w, "index.html", artists)
}

func HandleArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		RenderError(w, http.StatusBadRequest, "400 Bad Request: Missing artist ID")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		RenderError(w, http.StatusBadRequest, "400 Bad Request: Invalid artist ID")
		return
	}

	artistDetail, err := GetArtistDetail(id)
	if err != nil {
		if errors.Is(err, ErrArtistNotFound) {
			RenderError(w, http.StatusNotFound, fmt.Sprintf("404 Artist with ID %d Not Found", id))
			return
		}
		log.Printf("Error getting artist detail for ID %d: %v", id, err)
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	RenderPage(w, "artist.html", artistDetail)
}
