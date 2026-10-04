package helpers

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

var testArtists = []Artist{
	{ID: 1, Name: "Queen", Members: []string{"Freddie Mercury", "Brian May"}, CreationDate: 1970, FirstAlbum: "14-12-1973"},
	{ID: 2, Name: "Pink Floyd", Members: []string{"Roger Waters"}, CreationDate: 1965, FirstAlbum: "05-08-1967"},
}
var testRelations = []Relation{
	{ID: 1, DatesLocations: map[string][]string{"london-uk": {"09-12-2019"}}},
	{ID: 2, DatesLocations: map[string][]string{"north_carolina-usa": {"23-08-2019"}}},
}

// useTestSetup makes the tests work without the internet and finds the templates.
func useTestSetup(t *testing.T) {
	savedArtists, savedRelations, savedAt = testArtists, testRelations, time.Now()
	old := TemplatesDir
	TemplatesDir = filepath.Join("..", "web", "templates")
	t.Cleanup(func() { TemplatesDir = old })
}

func TestSearch(t *testing.T) {
	tests := []struct {
		query    string
		wantType string
		wantID   int
		wantLen  int
	}{
		{"queen", "artist", 1, 1},
		{"QUEEN", "artist", 1, 1},
		{"freddie", "member", 1, 1},
		{"north carolina", "location", 2, 1},
		{"london", "location", 1, 1},
		{"14-12-1973", "firstAlbum", 1, 1},
		{"1965", "creationDate", 2, 1},
		{"zzzz", "", 0, 0},
		{"", "", 0, 0},
	}
	for _, tt := range tests {
		got := Search(tt.query, testArtists, testRelations)
		if len(got) != tt.wantLen {
			t.Errorf("%q: got %d results, want %d", tt.query, len(got), tt.wantLen)
			continue
		}
		if tt.wantLen > 0 && (got[0].Type != tt.wantType || got[0].ID != tt.wantID) {
			t.Errorf("%q: got %+v", tt.query, got[0])
		}
	}
}

func TestSearchLimit(t *testing.T) {
	var many []Artist
	for i := 1; i <= 60; i++ {
		many = append(many, Artist{ID: i, Name: "Band " + strconv.Itoa(i)})
	}
	if got := Search("band", many, nil); len(got) != 50 {
		t.Errorf("expected 50 results, got %d", len(got))
	}
}

func TestHandleSearch(t *testing.T) {
	useTestSetup(t)

	tests := []struct {
		name   string
		method string
		url    string
		status int
		body   string
	}{
		{"match", http.MethodGet, "/search?q=queen", 200, "Queen (Artist)"},
		{"location", http.MethodGet, "/search?q=london", 200, "London, UK"},
		{"no match", http.MethodGet, "/search?q=zzzz", 200, "No results found"},
		{"empty query", http.MethodGet, "/search?q=", 200, "Type a band"},
		{"html is escaped", http.MethodGet, "/search?q=%3Cscript%3E", 200, "&lt;script&gt;"},
		{"POST rejected", http.MethodPost, "/search?q=a", 405, "405"},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		HandleSearch(rec, httptest.NewRequest(tt.method, tt.url, nil))
		if rec.Code != tt.status || !strings.Contains(rec.Body.String(), tt.body) {
			t.Errorf("%s: status=%d, body has %q: %v", tt.name, rec.Code, tt.body, strings.Contains(rec.Body.String(), tt.body))
		}
	}
}
