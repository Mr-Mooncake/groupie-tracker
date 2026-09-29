package helpers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupHandlerTest(t *testing.T) {
	oldCache, oldDir := searchCache, templatesDir
	searchCache, templatesDir = newTestCache(nil), "../templates"
	t.Cleanup(func() { searchCache, templatesDir = oldCache, oldDir })
}

func TestPages(t *testing.T) {
	setupHandlerTest(t)

	tests := []struct {
		name    string
		method  string
		target  string
		handler http.HandlerFunc
		status  int
		body    string
	}{
		{"home ok", "GET", "/", HandleHome, 200, "Queen"},
		{"home unknown path", "GET", "/nonexistent", HandleHome, 404, "Page not found"},
		{"home wrong method", "POST", "/", HandleHome, 405, ""},
		{"artist ok", "GET", "/artist?id=1", HandleArtist, 200, "London, UK"},
		{"artist not found", "GET", "/artist?id=9999", HandleArtist, 404, "Artist not found"},
		{"artist bad id", "GET", "/artist?id=abc", HandleArtist, 400, "Invalid artist id"},
		{"artist missing id", "GET", "/artist", HandleArtist, 400, ""},
		{"artist wrong method", "POST", "/artist?id=1", HandleArtist, 405, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tt.handler(rec, httptest.NewRequest(tt.method, tt.target, nil))
			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d", rec.Code, tt.status)
			}
			if !strings.Contains(rec.Body.String(), tt.body) {
				t.Errorf("body missing %q", tt.body)
			}
		})
	}
}
