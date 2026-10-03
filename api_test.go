package main

import (
	"testing"
)

func TestFetchAllArtists(t *testing.T) {
	artists, err := FetchAllArtists()
	if err != nil {
		t.Fatalf("Expected no error fetching artists, got: %v", err)
	}
	if len(artists) == 0 {
		t.Fatalf("Expected artists slice to be non-empty")
	}
}

func TestGetArtistDetail(t *testing.T) {
	detail, err := GetArtistDetail(1)
	if err != nil {
		t.Fatalf("Expected no error fetching artist detail for ID 1, got: %v", err)
	}
	if detail.Artist.Name != "Queen" {
		t.Errorf("Expected artist name 'Queen', got '%s'", detail.Artist.Name)
	}
	if len(detail.Relations) == 0 {
		t.Errorf("Expected relation data for Queen, got empty map")
	}
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
		result := formatLocation(tt.input)
		if result != tt.expected {
			t.Errorf("formatLocation(%q) = %q; expected %q", tt.input, result, tt.expected)
		}
	}
}
