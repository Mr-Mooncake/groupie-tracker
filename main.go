package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

func main() {
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/artist", handleArtist)

	port := ":8080"
	fmt.Printf("🚀 Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		renderError(w, http.StatusNotFound, "404 Page Not Found")
		return
	}

	if r.Method != http.MethodGet {
		renderError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}

	artists, err := FetchAllArtists()
	if err != nil {
		log.Printf("Error fetching artists: %v", err)
		renderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "index.html"))
	if err != nil {
		log.Printf("Error parsing template index.html: %v", err)
		renderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	err = tmpl.Execute(w, artists)
	if err != nil {
		log.Printf("Error executing template index.html: %v", err)
		renderError(w, http.StatusInternalServerError, "500 Internal Server Error")
	}
}

func handleArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		renderError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		renderError(w, http.StatusBadRequest, "400 Bad Request: Missing artist ID")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		renderError(w, http.StatusBadRequest, "400 Bad Request: Invalid artist ID")
		return
	}

	artistDetail, err := GetArtistDetail(id)
	if err != nil {
		if errors.Is(err, ErrArtistNotFound) {
			renderError(w, http.StatusNotFound, fmt.Sprintf("404 Artist with ID %d Not Found", id))
			return
		}
		log.Printf("Error getting artist detail for ID %d: %v", id, err)
		renderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "artist.html"))
	if err != nil {
		log.Printf("Error parsing template artist.html: %v", err)
		renderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	err = tmpl.Execute(w, artistDetail)
	if err != nil {
		log.Printf("Error executing template artist.html: %v", err)
		renderError(w, http.StatusInternalServerError, "500 Internal Server Error")
	}
}

func renderError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)

	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "error.html"))
	if err != nil {
		http.Error(w, message, statusCode)
		return
	}

	data := ErrorData{
		StatusCode: statusCode,
		Message:    message,
	}

	err = tmpl.Execute(w, data)
}
