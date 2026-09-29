package helpers

// [Yasser] Client-server event: live search endpoint + data cache.

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	apiBase         = "https://groupietrackers.herokuapp.com/api"
	maxSearchResult = 10
	cacheTTL        = 10 * time.Minute
)

// SearchResult is one suggestion sent to the browser.
type SearchResult struct {
	Text string `json:"text"`
	Type string `json:"type"`
	ID   int    `json:"id"`
}

// ---------- cache ----------

// dataCache keeps artists + relations in memory so we don't call the
// remote API on every request. Safe for concurrent use.
type dataCache struct {
	mu        sync.Mutex
	artists   []Artist
	relations []Relation
	fetchedAt time.Time
	ttl       time.Duration
	load      func() ([]Artist, []Relation, error)
}

func (c *dataCache) get() ([]Artist, []Relation, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.artists != nil && time.Since(c.fetchedAt) < c.ttl {
		return c.artists, c.relations, nil
	}

	artists, relations, err := c.load()
	if err != nil {
		if c.artists != nil { // serve stale data rather than fail
			return c.artists, c.relations, nil
		}
		return nil, nil, err
	}

	c.artists, c.relations, c.fetchedAt = artists, relations, time.Now()
	return artists, relations, nil
}

func loadFromAPI() ([]Artist, []Relation, error) {
	var artists []Artist
	if err := fetchJSON(apiBase+"/artists", &artists); err != nil {
		return nil, nil, err
	}
	var idx RelationIndex
	if err := fetchJSON(apiBase+"/relation", &idx); err != nil {
		return nil, nil, err
	}
	return artists, idx.Index, nil
}

var searchCache = &dataCache{ttl: cacheTTL, load: loadFromAPI}

// ---------- search logic (pure, easy to test) ----------

// searchIndex returns up to limit suggestions, grouped by category so
// artist-name matches come first.
func searchIndex(query string, artists []Artist, relations []Relation, limit int) []SearchResult {
	results := []SearchResult{}
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return results
	}

	match := func(s string) bool { return strings.Contains(strings.ToLower(s), q) }
	add := func(text, typ string, id int) bool {
		results = append(results, SearchResult{Text: text, Type: typ, ID: id})
		return len(results) >= limit
	}

	for _, a := range artists {
		if match(a.Name) && add(a.Name+" (Artist)", "artist", a.ID) {
			return results
		}
	}
	for _, a := range artists {
		for _, m := range a.Members {
			if match(m) && add(m+" - Member ("+a.Name+")", "member", a.ID) {
				return results
			}
		}
	}

	names := make(map[int]string, len(artists))
	for _, a := range artists {
		names[a.ID] = a.Name
	}
	for _, r := range relations {
		name, ok := names[r.ID]
		if !ok {
			continue
		}
		keys := make([]string, 0, len(r.DatesLocations))
		for k := range r.DatesLocations {
			keys = append(keys, k)
		}
		sort.Strings(keys) // deterministic order
		for _, k := range keys {
			pretty := FormatLocation(k)
			if (match(pretty) || match(k)) && add(pretty+" - Location ("+name+")", "location", r.ID) {
				return results
			}
		}
	}

	for _, a := range artists {
		if match(a.FirstAlbum) && add(a.FirstAlbum+" - First Album ("+a.Name+")", "firstAlbum", a.ID) {
			return results
		}
	}
	for _, a := range artists {
		year := strconv.Itoa(a.CreationDate)
		if match(year) && add(year+" - Created ("+a.Name+")", "creationDate", a.ID) {
			return results
		}
	}
	return results
}

// ---------- HTTP handler ----------

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func newSearchHandler(c func() *dataCache) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if query == "" {
			json.NewEncoder(w).Encode([]SearchResult{})
			return
		}

		artists, relations, err := c().get()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not load data")
			return
		}

		json.NewEncoder(w).Encode(searchIndex(query, artists, relations, maxSearchResult))
	}
}

// HandleSearch serves GET /api/search?q=...
var HandleSearch = newSearchHandler(func() *dataCache { return searchCache })
