package helpers

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

// HandleHome answers GET / with the homepage and artist list.
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

// HandleArtist answers GET /artist?id=... with artist details and concert dates.
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

// HandleSearch answers GET /search?q=... by showing a results page.
func HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}
	artists, relations, err := GetData()
	if err != nil {
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	data := struct {
		Query   string
		Results []SearchResult
	}{query, Search(query, artists, relations)}

	RenderPage(w, "search.html", data)
}

// HandleRandom answers GET /random by redirecting to a random artist page.
func HandleRandom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}

	artists, _, err := GetData()
	if err != nil || len(artists) == 0 {
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	randomArtist := artists[rand.Intn(len(artists))]
	http.Redirect(w, r, "/artist?id="+strconv.Itoa(randomArtist.ID), http.StatusSeeOther)
}
