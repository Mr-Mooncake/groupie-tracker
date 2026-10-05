package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"groupie-tracker/helpers"
)

var testArtists = []helpers.Artist{
	{ID: 1, Name: "Queen", Members: []string{"Freddie Mercury", "Brian May"}, CreationDate: 1970, FirstAlbum: "14-12-1973"},
	{ID: 2, Name: "Pink Floyd", Members: []string{"Roger Waters"}, CreationDate: 1965, FirstAlbum: "05-08-1967"},
}

var testRelations = []helpers.Relation{
	{ID: 1, DatesLocations: map[string][]string{"london-uk": {"09-12-2019"}}},
	{ID: 2, DatesLocations: map[string][]string{"north_carolina-usa": {"23-08-2019"}}},
}

func setupTestEnv(t *testing.T) {
	helpers.SetCachedData(testArtists, testRelations)
	old := helpers.TemplatesDir
	helpers.TemplatesDir = filepath.Join("web", "templates")
	t.Cleanup(func() { helpers.TemplatesDir = old })
}

func TestFormatLocation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"london-uk", "London, UK"},
		{"north_carolina-usa", "North Carolina, USA"},
		{"sao_paulo-brazil", "Sao Paulo, Brazil"},
	}

	for _, tt := range tests {
		result := helpers.FormatLocation(tt.input)
		if result != tt.expected {
			t.Errorf("FormatLocation(%q) = %q; expected %q", tt.input, result, tt.expected)
		}
	}
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
		got := helpers.Search(tt.query, testArtists, testRelations)
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
	var many []helpers.Artist
	for i := 1; i <= 60; i++ {
		many = append(many, helpers.Artist{ID: i, Name: "Band " + strconv.Itoa(i)})
	}
	if got := helpers.Search("band", many, nil); len(got) != 50 {
		t.Errorf("expected 50 results, got %d", len(got))
	}
}

func TestHandleSearch(t *testing.T) {
	setupTestEnv(t)

	tests := []struct {
		name   string
		method string
		url    string
		status int
		body   string
	}{
		{"match", http.MethodGet, "/search?q=queen", http.StatusOK, "Queen (Artist)"},
		{"location", http.MethodGet, "/search?q=london", http.StatusOK, "London, UK"},
		{"no match", http.MethodGet, "/search?q=zzzz", http.StatusOK, "No results found"},
		{"empty query", http.MethodGet, "/search?q=", http.StatusOK, "Type a band"},
		{"html is escaped", http.MethodGet, "/search?q=%3Cscript%3E", http.StatusOK, "&lt;script&gt;"},
		{"POST rejected", http.MethodPost, "/search?q=a", http.StatusMethodNotAllowed, "405"},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		helpers.HandleSearch(rec, httptest.NewRequest(tt.method, tt.url, nil))
		if rec.Code != tt.status || !strings.Contains(rec.Body.String(), tt.body) {
			t.Errorf("%s: status=%d, body has %q: %v", tt.name, rec.Code, tt.body, strings.Contains(rec.Body.String(), tt.body))
		}
	}
}

func TestHandleRandom(t *testing.T) {
	setupTestEnv(t)

	// Several tries, because the choice is random.
	for i := 0; i < 20; i++ {
		rec := httptest.NewRecorder()
		helpers.HandleRandom(rec, httptest.NewRequest(http.MethodGet, "/random", nil))
		loc := rec.Header().Get("Location")
		if rec.Code != http.StatusSeeOther || (loc != "/artist?id=1" && loc != "/artist?id=2") {
			t.Fatalf("got status %d, location %q", rec.Code, loc)
		}
	}

	rec := httptest.NewRecorder()
	helpers.HandleRandom(rec, httptest.NewRequest(http.MethodPost, "/random", nil))
	if rec.Code != http.StatusMethodNotAllowed || !strings.Contains(rec.Body.String(), "405") {
		t.Errorf("POST: got status %d", rec.Code)
	}
}

func TestHandleHome(t *testing.T) {
	setupTestEnv(t)

	// Test 404 for wrong path
	rec := httptest.NewRecorder()
	HandleHome(rec, httptest.NewRequest(http.MethodGet, "/invalid", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for invalid path, got %d", rec.Code)
	}

	// Test 405 for POST
	rec = httptest.NewRecorder()
	HandleHome(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST, got %d", rec.Code)
	}
}

func TestHandleArtist(t *testing.T) {
	setupTestEnv(t)

	// Test 405 for POST
	rec := httptest.NewRecorder()
	HandleArtist(rec, httptest.NewRequest(http.MethodPost, "/artist?id=1", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST, got %d", rec.Code)
	}

	// Test 400 for missing ID
	rec = httptest.NewRecorder()
	HandleArtist(rec, httptest.NewRequest(http.MethodGet, "/artist", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing ID, got %d", rec.Code)
	}

	// Test 400 for invalid ID
	rec = httptest.NewRecorder()
	HandleArtist(rec, httptest.NewRequest(http.MethodGet, "/artist?id=abc", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid ID, got %d", rec.Code)
	}

	// Test 400 for non-positive ID
	rec = httptest.NewRecorder()
	HandleArtist(rec, httptest.NewRequest(http.MethodGet, "/artist?id=-5", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for negative ID, got %d", rec.Code)
	}
}

func TestFetchAllArtists(t *testing.T) {
	artists, err := helpers.FetchAllArtists()
	if err != nil {
		t.Skipf("Skipping API test due to network unavailability: %v", err)
	}
	if len(artists) == 0 {
		t.Fatalf("Expected artists slice to be non-empty")
	}
}

func TestGetArtistDetail(t *testing.T) {
	detail, err := helpers.GetArtistDetail(1)
	if err != nil {
		t.Skipf("Skipping API test due to network unavailability: %v", err)
	}
	if detail.Artist.Name != "Queen" {
		t.Errorf("Expected artist name 'Queen', got '%s'", detail.Artist.Name)
	}
	if len(detail.Relations) == 0 {
		t.Errorf("Expected relation data for Queen, got empty map")
	}
}
