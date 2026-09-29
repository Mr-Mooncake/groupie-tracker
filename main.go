package main

import (
	"fmt"
	"net/http"

	"groupie-tracker/helpers"
)

func main() {
	http.HandleFunc("/", helpers.HandleHome)
	http.HandleFunc("/artist", helpers.HandleArtist)
	http.HandleFunc("/api/search", helpers.HandleSearch)

	// serves /map/script.js and /map/style.css
	http.Handle("/map/", http.StripPrefix("/map/", http.FileServer(http.Dir("map"))))

	fmt.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
