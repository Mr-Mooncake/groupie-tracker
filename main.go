package main

import (
	"fmt"
	"groupie-tracker/helpers"
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", helpers.HandleHome)
	http.HandleFunc("/artist", helpers.HandleArtist)
	http.HandleFunc("/search", helpers.HandleSearch) // [Yasser] search results page
	http.HandleFunc("/random", helpers.HandleRandom) // [Yasser] random artist button

	port := ":8080"
	fmt.Printf("🚀 Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
