package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"groupie-tracker/helpers"
)

func main() {
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", HandleHome)
	http.HandleFunc("/artist", HandleArtist)
	http.HandleFunc("/api/search", helpers.HandleSearch) // [Yasser] live search endpoint

	port := ":8080"
	fmt.Printf("🚀 Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
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

	artists, err := helpers.FetchAllArtists()
	if err != nil {
		log.Printf("Error fetching artists: %v", err)
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "index.html"))
	if err != nil {
		log.Printf("Error parsing template index.html: %v", err)
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	err = tmpl.Execute(w, artists)
	if err != nil {
		log.Printf("Error executing template index.html: %v", err)
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
	}
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

	artistDetail, err := helpers.GetArtistDetail(id)
	if err != nil {
		if errors.Is(err, helpers.ErrArtistNotFound) {
			RenderError(w, http.StatusNotFound, fmt.Sprintf("404 Artist with ID %d Not Found", id))
			return
		}
		log.Printf("Error getting artist detail for ID %d: %v", id, err)
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "artist.html"))
	if err != nil {
		log.Printf("Error parsing template artist.html: %v", err)
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	err = tmpl.Execute(w, artistDetail)
	if err != nil {
		log.Printf("Error executing template artist.html: %v", err)
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
	}
}

func RenderError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)

	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "error.html"))
	if err != nil {
		http.Error(w, message, statusCode)
		return
	}

	data := helpers.ErrorData{
		StatusCode: statusCode,
		Message:    message,
	}

	err = tmpl.Execute(w, data)
}
