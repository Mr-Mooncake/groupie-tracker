package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const APIBaseURL = "https://groupietrackers.herokuapp.com/api"

var ErrArtistNotFound = errors.New("artist not found")

func fetchJSON(url string, target any) error {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("network error fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code %d from %s", resp.StatusCode, url)
	}

	err = json.NewDecoder(resp.Body).Decode(target)
	if err != nil {
		return fmt.Errorf("failed to decode JSON from %s: %v", url, err)
	}

	return nil
}

func FetchAllArtists() ([]Artist, error) {
	var artists []Artist
	err := fetchJSON(APIBaseURL+"/artists", &artists)
	if err != nil {
		return nil, err
	}
	return artists, nil
}

func GetArtistDetail(id int) (ArtistDetail, error) {
	artists, err := FetchAllArtists()
	if err != nil {
		return ArtistDetail{}, err
	}

	var foundArtist Artist
	found := false
	for _, a := range artists {
		if a.ID == id {
			foundArtist = a
			found = true
			break
		}
	}
	if !found {
		return ArtistDetail{}, ErrArtistNotFound
	}

	var relIndex RelationIndex
	err = fetchJSON(APIBaseURL+"/relation", &relIndex)
	if err != nil {
		return ArtistDetail{}, err
	}

	formattedRelations := make(map[string][]string)
	for _, r := range relIndex.Index {
		if r.ID == id {
			for loc, dates := range r.DatesLocations {
				cleanLoc := formatLocation(loc)
				formattedRelations[cleanLoc] = dates
			}
			break
		}
	}

	return ArtistDetail{
		Artist:    foundArtist,
		Relations: formattedRelations,
	}, nil
}

func formatLocation(raw string) string {
	parts := strings.Split(raw, "-")
	for i, part := range parts {
		part = strings.ReplaceAll(part, "_", " ")
		words := strings.Fields(part)
		for j, word := range words {
			if strings.EqualFold(word, "usa") || strings.EqualFold(word, "uk") {
				words[j] = strings.ToUpper(word)
			} else if len(word) > 0 {
				words[j] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
			}
		}
		parts[i] = strings.Join(words, " ")
	}
	return strings.Join(parts, ", ")
}
