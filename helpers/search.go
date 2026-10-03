package helpers

// [Yasser] Live search: the browser asks /api/search?q=... and gets a JSON list back.

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// One suggestion in the dropdown (the json tags are the names JavaScript sees).
type SearchResult struct {
	Text string `json:"text"`
	Type string `json:"type"`
	ID   int    `json:"id"`
}

// Saved copy of the API data (a "cache"), so we don't call the API on every
// keystroke. The mutex stops two requests from changing it at the same time.
var (
	mutex          sync.Mutex
	savedArtists   []Artist
	savedRelations []Relation
	savedAt        time.Time
)

// getData returns artists and relations, from memory when possible.
func GetData() ([]Artist, []Relation, error) {
	mutex.Lock()
	defer mutex.Unlock() // "defer" = run this when the function ends

	if savedArtists != nil && time.Since(savedAt) < 10*time.Minute {
		return savedArtists, savedRelations, nil // saved copy is still fresh
	}

	artists, err1 := FetchAllArtists()
	var index RelationIndex
	err2 := FetchJSON(APIBaseURL+"/relation", &index)
	if err1 != nil || err2 != nil {
		if savedArtists != nil { // API is down: old data is better than nothing
			return savedArtists, savedRelations, nil
		}
		return nil, nil, errors.New("could not load data")
	}

	savedArtists, savedRelations, savedAt = artists, index.Index, time.Now()
	return savedArtists, savedRelations, nil
}

// contains checks if text includes query, ignoring upper/lower case.
func Contains(text, query string) bool {
	return strings.Contains(strings.ToLower(text), query)
}

// search returns up to 10 results that match the query.
func Search(query string, artists []Artist, relations []Relation) []SearchResult {
	query = strings.ToLower(strings.TrimSpace(query))
	results := []SearchResult{} // empty list, so JSON gives [] and not null
	if query == "" {
		return results
	}
	for _, a := range artists { // 1. artist names
		if Contains(a.Name, query) {
			results = append(results, SearchResult{a.Name + " (Artist)", "artist", a.ID})
		}
	}
	for _, a := range artists { // 2. band members
		for _, member := range a.Members {
			if Contains(member, query) {
				results = append(results, SearchResult{member + " - Member (" + a.Name + ")", "member", a.ID})
			}
		}
	}

	names := map[int]string{} // id -> artist name, needed for step 3
	for _, a := range artists {
		names[a.ID] = a.Name
	}
	for _, r := range relations { // 3. concert locations
		var places []string
		for place := range r.DatesLocations {
			places = append(places, place)
		}
		sort.Strings(places) // map order is random; sorting keeps results stable
		for _, place := range places {
			nice := FormatLocation(place) // "london-uk" -> "London, UK"
			if Contains(nice, query) || Contains(place, query) {
				results = append(results, SearchResult{nice + " - Location (" + names[r.ID] + ")", "location", r.ID})
			}
		}
	}
	for _, a := range artists { // 4. first album date and creation year
		year := strconv.Itoa(a.CreationDate) // number -> text
		if Contains(a.FirstAlbum, query) {
			results = append(results, SearchResult{a.FirstAlbum + " - First Album (" + a.Name + ")", "firstAlbum", a.ID})
		}
		if Contains(year, query) {
			results = append(results, SearchResult{year + " - Created (" + a.Name + ")", "creationDate", a.ID})
		}
	}

	if len(results) > 10 { // keep the dropdown short
		results = results[:10]
	}
	return results
}

// handleSearch answers GET /api/search?q=...
func HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	artists, relations, err := GetData()
	if err != nil {
		http.Error(w, "could not load data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Search(r.URL.Query().Get("q"), artists, relations))
}
