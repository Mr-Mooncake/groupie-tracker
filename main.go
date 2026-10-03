package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
<<<<<<< HEAD

	"groupie-tracker/helpers"
=======
>>>>>>> origin/main
)

func main() {
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

<<<<<<< HEAD
	http.HandleFunc("/", HandleHome)
	http.HandleFunc("/artist", HandleArtist)
	http.HandleFunc("/api/search", helpers.HandleSearch) // [Yasser] live search endpoint
=======
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/artist", handleArtist)
>>>>>>> origin/main

	port := ":8080"
	fmt.Printf("🚀 Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

<<<<<<< HEAD
func HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		RenderError(w, http.StatusNotFound, "404 Page Not Found")
=======
func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		renderError(w, http.StatusNotFound, "404 Page Not Found")
>>>>>>> origin/main
		return
	}

	if r.Method != http.MethodGet {
<<<<<<< HEAD
		RenderError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}

	artists, err := helpers.FetchAllArtists()
	if err != nil {
		log.Printf("Error fetching artists: %v", err)
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
=======
		renderError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}

	artists, err := FetchAllArtists()
	if err != nil {
		log.Printf("Error fetching artists: %v", err)
		renderError(w, http.StatusInternalServerError, "500 Internal Server Error")
>>>>>>> origin/main
		return
	}

	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "index.html"))
	if err != nil {
		log.Printf("Error parsing template index.html: %v", err)
<<<<<<< HEAD
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
=======
		renderError(w, http.StatusInternalServerError, "500 Internal Server Error")
>>>>>>> origin/main
		return
	}

	err = tmpl.Execute(w, artists)
	if err != nil {
		log.Printf("Error executing template index.html: %v", err)
<<<<<<< HEAD
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
	}
}

func HandleArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
=======
		renderError(w, http.StatusInternalServerError, "500 Internal Server Error")
	}
}

func handleArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		renderError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
>>>>>>> origin/main
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
<<<<<<< HEAD
		RenderError(w, http.StatusBadRequest, "400 Bad Request: Missing artist ID")
=======
		renderError(w, http.StatusBadRequest, "400 Bad Request: Missing artist ID")
>>>>>>> origin/main
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
<<<<<<< HEAD
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
=======
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
>>>>>>> origin/main
		return
	}

	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "artist.html"))
	if err != nil {
		log.Printf("Error parsing template artist.html: %v", err)
<<<<<<< HEAD
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
=======
		renderError(w, http.StatusInternalServerError, "500 Internal Server Error")
>>>>>>> origin/main
		return
	}

	err = tmpl.Execute(w, artistDetail)
	if err != nil {
		log.Printf("Error executing template artist.html: %v", err)
<<<<<<< HEAD
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
	}
}

func RenderError(w http.ResponseWriter, statusCode int, message string) {
=======
		renderError(w, http.StatusInternalServerError, "500 Internal Server Error")
	}
}

func renderError(w http.ResponseWriter, statusCode int, message string) {
>>>>>>> origin/main
	w.WriteHeader(statusCode)

	tmpl, err := template.ParseFiles(filepath.Join("web", "templates", "error.html"))
	if err != nil {
		http.Error(w, message, statusCode)
		return
	}

<<<<<<< HEAD
	data := helpers.ErrorData{
=======
	data := ErrorData{
>>>>>>> origin/main
		StatusCode: statusCode,
		Message:    message,
	}

	err = tmpl.Execute(w, data)
}
