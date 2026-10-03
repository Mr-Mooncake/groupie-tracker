package helpers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
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
	for i := 1; i <= 15; i++ {
		many = append(many, Artist{ID: i, Name: "Band " + strconv.Itoa(i)})
	}
	if got := Search("band", many, nil); len(got) != 10 {
		t.Errorf("expected 10 results, got %d", len(got))
	}
}

func TestHandleSearch(t *testing.T) {
	// Fill the saved copy so the test does not need the internet.
	savedArtists, savedRelations, savedAt = testArtists, testRelations, time.Now()

	rec := httptest.NewRecorder()
	HandleSearch(rec, httptest.NewRequest(http.MethodGet, "/api/search?q=queen", nil))
	if rec.Code != 200 {
		t.Errorf("GET: got status %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	HandleSearch(rec, httptest.NewRequest(http.MethodGet, "/api/search?q=", nil))
	if rec.Body.String() != "[]\n" {
		t.Errorf("empty query: got %q, want []", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	HandleSearch(rec, httptest.NewRequest(http.MethodPost, "/api/search?q=a", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST: got status %d, want 405", rec.Code)
	}
}
