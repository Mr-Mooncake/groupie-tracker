package helpers

// [Yasser] Search: the browser submits a form to /search?q=... and gets a results page back.

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// One search result: the text to show, what kind it is, and the band's id (for the link).
type SearchResult struct {
	Text string
	Type string
	ID   int
}

// Saved copy of the API data (a "cache"), so we don't call the API on every
// search. The mutex stops two requests from changing it at the same time.
var (
	mutex          sync.Mutex
	savedArtists   []Artist
	savedRelations []Relation
	savedAt        time.Time
)

// GetData returns artists and relations, from memory when possible.
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

// SetCachedData sets the in-memory cache, useful for tests or offline data.
func SetCachedData(artists []Artist, relations []Relation) {
	mutex.Lock()
	defer mutex.Unlock()
	savedArtists, savedRelations, savedAt = artists, relations, time.Now()
}

// Contains checks if text includes query, ignoring upper/lower case.
func Contains(text, query string) bool {
	return strings.Contains(strings.ToLower(text), query)
}

// Search returns up to 50 results that match the query.
func Search(query string, artists []Artist, relations []Relation) []SearchResult {
	query = strings.ToLower(strings.TrimSpace(query))
	results := []SearchResult{}
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

	if len(results) > 50 { // keep the page short
		results = results[:50]
	}
	return results
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

	query := strings.TrimSpace(r.URL.Query().Get("q")) // the text after ?q=
	data := struct {
		Query   string
		Results []SearchResult
	}{query, Search(query, artists, relations)}

	RenderPage(w, "search.html", data)
}
