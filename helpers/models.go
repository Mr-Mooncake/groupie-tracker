package helpers

// [Hashim] Data models and the generic fetch function.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Artist holds basic band information from /api/artists.
type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Relations    string   `json:"relations"`
}

// Relation maps concert locations to dates for one artist.
type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// RelationIndex is the wrapper object returned by /api/relation.
type RelationIndex struct {
	Index []Relation `json:"index"`
}

// ArtistDetail combines Artist and Relation data for the artist page.
type ArtistDetail struct {
	Artist    Artist
	Relations map[string][]string
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// fetchJSON sends a GET request and decodes the JSON body into target.
func fetchJSON(url string, target interface{}) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("network error fetching %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d from %s", resp.StatusCode, url)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode JSON from %s: %v", url, err)
	}
	return nil
}
