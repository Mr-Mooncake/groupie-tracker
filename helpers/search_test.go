package helpers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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

func newTestCache(err error) *dataCache {
	return &dataCache{ttl: time.Minute, load: func() ([]Artist, []Relation, error) {
		if err != nil {
			return nil, nil, err
		}
		return testArtists, testRelations, nil
	}}
}

func TestSearchIndex(t *testing.T) {
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
		got := searchIndex(tt.query, testArtists, testRelations, 10)
		if got == nil {
			t.Fatalf("%q: got nil slice, want empty slice (JSON must be [])", tt.query)
		}
		if len(got) != tt.wantLen {
			t.Errorf("%q: got %d results, want %d", tt.query, len(got), tt.wantLen)
			continue
		}
		if tt.wantLen > 0 && (got[0].Type != tt.wantType || got[0].ID != tt.wantID) {
			t.Errorf("%q: got %+v, want type=%s id=%d", tt.query, got[0], tt.wantType, tt.wantID)
		}
	}
}

func TestSearchIndexLimit(t *testing.T) {
	if got := searchIndex("e", testArtists, testRelations, 3); len(got) != 3 {
		t.Errorf("limit not applied: got %d", len(got))
	}
}

func TestSearchHandler(t *testing.T) {
	c := newTestCache(nil)
	h := newSearchHandler(func() *dataCache { return c })

	t.Run("match", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h(rec, httptest.NewRequest(http.MethodGet, "/api/search?q=queen", nil))
		var res []SearchResult
		if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &res) != nil || len(res) != 1 {
			t.Fatalf("bad response: %d %s", rec.Code, rec.Body.String())
		}
	})
	t.Run("empty query returns []", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h(rec, httptest.NewRequest(http.MethodGet, "/api/search?q=", nil))
		if rec.Code != 200 || rec.Body.String() != "[]\n" {
			t.Fatalf("got %d %q", rec.Code, rec.Body.String())
		}
	})
	t.Run("POST rejected", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h(rec, httptest.NewRequest(http.MethodPost, "/api/search?q=a", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("got %d", rec.Code)
		}
	})
	t.Run("upstream failure returns 500", func(t *testing.T) {
		bad := newTestCache(errors.New("boom"))
		hb := newSearchHandler(func() *dataCache { return bad })
		rec := httptest.NewRecorder()
		hb(rec, httptest.NewRequest(http.MethodGet, "/api/search?q=queen", nil))
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("got %d", rec.Code)
		}
	})
}

func TestCacheServesStaleOnError(t *testing.T) {
	calls := 0
	c := &dataCache{ttl: time.Nanosecond, load: func() ([]Artist, []Relation, error) {
		calls++
		if calls > 1 {
			return nil, nil, errors.New("down")
		}
		return testArtists, testRelations, nil
	}}
	c.get()
	time.Sleep(time.Millisecond)
	if a, _, err := c.get(); err != nil || len(a) == 0 {
		t.Fatalf("expected stale data, got err=%v", err)
	}
}
