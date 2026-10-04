package helpers

// [Yasser] Random artist button: /random picks an artist and redirects to its page.

import (
	"math/rand"
	"net/http"
	"strconv"
)

// HandleRandom answers GET /random.
func HandleRandom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}

	artists, _, err := GetData() // the saved list (slice) of all artists
	if err != nil || len(artists) == 0 {
		RenderError(w, http.StatusInternalServerError, "500 Internal Server Error")
		return
	}

	// Pick a random position in the list, then use that artist's ID.
	// This way the ID is always between 1 and the last ID, and always exists.
	randomArtist := artists[rand.Intn(len(artists))]
	http.Redirect(w, r, "/artist?id="+strconv.Itoa(randomArtist.ID), http.StatusSeeOther)
}
