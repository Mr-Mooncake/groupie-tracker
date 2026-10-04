package helpers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleRandom(t *testing.T) {
	useTestSetup(t)

	// Several tries, because the choice is random.
	for i := 0; i < 20; i++ {
		rec := httptest.NewRecorder()
		HandleRandom(rec, httptest.NewRequest(http.MethodGet, "/random", nil))
		loc := rec.Header().Get("Location")
		if rec.Code != http.StatusSeeOther || (loc != "/artist?id=1" && loc != "/artist?id=2") {
			t.Fatalf("got status %d, location %q", rec.Code, loc)
		}
	}

	rec := httptest.NewRecorder()
	HandleRandom(rec, httptest.NewRequest(http.MethodPost, "/random", nil))
	if rec.Code != http.StatusMethodNotAllowed || !strings.Contains(rec.Body.String(), "405") {
		t.Errorf("POST: got status %d", rec.Code)
	}
}
