# Groupie Trackers

## Description

**Groupie Trackers** is a web application that fetches data from the public [Groupie Trackers API](https://groupietrackers.herokuapp.com/api) and displays it as a user-friendly website. Users can browse a gallery of bands and artists, open a detail page for each one, search the data, and jump to a random artist.

The project is a small full-stack Go application:

- A **Go HTTP server** (`net/http`) handles routing, validation, API calls, and rendering.
- **Go HTML templates** (`html/template`) render every page and safely escape any text coming from the API or from the user.
- Plain **HTML/CSS** provides the layout and styling. The site uses **no JavaScript**: every interaction is a normal request to the Go server.
- Data is fetched from the API's `artists` and `relation` endpoints. The `relation` endpoint already links each artist to its concert locations and dates, so the `locations` and `dates` endpoints are not called separately.

Only the Go standard library is used — no third-party packages.

### Features

| Feature | Description |
|---|---|
| Artist gallery | The home page shows every artist as a card with its picture and name. |
| Artist details | Each artist has a page with the picture, first album date, creation year, band members, and all concert locations with their dates. |
| Scrollable dates | Each concert location box shows five dates at a time; the rest are reached with a scroll bar. |
| Search | A search form looks through artist names, band members, concert locations, first album dates, and creation years. |
| Random artist | A button opens the page of a randomly chosen artist. |
| Error pages | Invalid routes, methods, IDs, and failures show a styled error page instead of crashing the server. |

### The client-server event

The project requires an event where an action of the client triggers a request to the server and the server responds. This project implements it in two ways, both without JavaScript:

1. **Search:** when the user submits the search form, the browser sends `GET /search?q=...`. The server searches the data and responds with a results page.
2. **Random artist:** when the user clicks the random button, the browser sends `GET /random`. The server picks an artist and responds with a redirect to that artist's page.

## Authors

- Yaser, Hussain, Sayed Hashim

| Member | Part |
|---|---|
| Hashim | Data models, API fetching, artist detail logic, routing, and error handling |
| Yaser | Search, random artist feature, and the client-server event |
| Hussain | HTML templates and CSS styling |

## Usage

### Prerequisites

- [Go](https://go.dev/dl/) installed (1.21+ recommended)
- An internet connection, because the data is fetched live from the Groupie Trackers API

### Project structure

```
groupie-tracker/
├── go.mod
├── main.go
├── main_test.go
├── helpers/
│   ├── api.go
│   ├── handlers.go
│   ├── models.go
│   ├── render.go
│   └── search.go
└── web/
    ├── static/
    │   └── style.css
    └── templates/
        ├── index.html
        ├── artist.html
        ├── search.html
        └── error.html
```

> Make sure the `web/` folder sits alongside `main.go` in the project root — the server reads the templates and static files using relative paths.

### Running the server

From the project root directory, run:

```bash
go run .
```

You should see:

```
🚀 Server running on http://localhost:8080
```

Then open your browser and navigate to:

```
http://localhost:8080
```

### Using the app

1. The home page lists all artists. Click a card to open that artist's page.
2. On an artist page, read the band information and scroll through the concert dates of each location.
3. Type a name, member, location, date, or year into the search box and press **Search** to see matching results. Click a result to open that artist.
4. Click the random/discover button to open a random artist.

### Endpoints

| Method | Route | Description |
|---|---|---|
| `GET` | `/` | Home page with the list of all artists. |
| `GET` | `/artist?id=<id>` | Detail page of one artist, with members and concert dates. |
| `GET` | `/search?q=<text>` | Results page for the search text. |
| `GET` | `/random` | Redirects to the page of a random artist. |
| `GET` | `/static/*` | Serves static assets (CSS). |

### HTTP status codes

- **200 OK** — request handled successfully.
- **303 See Other** — redirect from `/random` to the chosen artist page.
- **400 Bad Request** — the artist ID is missing, not a number, or not positive.
- **404 Not Found** — unknown route, or no artist with the requested ID.
- **405 Method Not Allowed** — any method other than `GET`.
- **500 Internal Server Error** — the API could not be reached, or a template failed to load or render.

## Testing

The project includes a Go test suite (`main_test.go`, in the project root) that covers the data logic and the HTTP handlers:

- **Location formatting:** raw API locations such as `north_carolina-usa` become `North Carolina, USA`.
- **Search:** matches on artist names, members, locations, first album dates, and years; case-insensitive matching; empty and non-matching queries; the 50-result limit.
- **Handlers:** status codes and page content for `/`, `/artist`, `/search`, and `/random`, including wrong methods, missing or invalid IDs, and HTML escaping of user input.
- **API calls:** fetching the artists and one artist's details from the real API. These tests are skipped automatically when there is no internet connection.

The handler tests load fake data into the cache, so they run without internet access.

To run the tests, from the project root:

```bash
go test ./... -v
```

Run the tests from the project root, since the templates are located using relative paths.

## Implementation details: algorithm

### 1. Fetching the data

`FetchJSON` sends an HTTP `GET` request with a 10-second timeout, checks for a `200 OK` status, and decodes the JSON body into Go structs (`Artist`, `Relation`, `RelationIndex`). A failure at any step returns an error instead of crashing the server.

The `artists` endpoint returns the band information (name, image, members, creation year, first album). The `relation` endpoint returns, for each artist ID, a map from concert locations to lists of dates.

### 2. Caching

Searching and the random feature need all the data, so `GetData` keeps a copy of the artists and relations in memory for 10 minutes:

1. If the saved copy is less than 10 minutes old, it is returned immediately.
2. Otherwise the API is called again and the copy is replaced.
3. If the API cannot be reached but an older copy exists, the older copy is used. If there is no copy at all, an error is returned.

A mutex makes sure that two requests never change the saved data at the same time.

### 3. Artist detail page

`HandleArtist` validates the `id` parameter (present, a number, greater than zero) and calls `GetArtistDetail`. This function finds the artist with that ID and its relation entry, and formats every location name with `FormatLocation`:

- the raw name is split at the `-`, for example `london-uk` into `london` and `uk`
- underscores become spaces and each word is capitalized
- `USA` and `UK` are written in capital letters

The result is `London, UK`. If no artist has that ID, the user gets a 404 page.

### 4. Search

`Search` lowercases and trims the query, then collects matches in this order, so artist names always appear first:

1. artist names
2. band members
3. concert locations (both the raw and the formatted name are checked)
4. first album dates and creation years

Each result stores the text to show, its type, and the artist ID used for the link. The list is limited to 50 results. `HandleSearch` passes the query and the results to `search.html`. An empty query shows a hint, and a query without matches shows a "No results found" message.

### 5. Random artist

`HandleRandom` takes the saved list of artists, picks a random position with `rand.Intn`, and redirects to `/artist?id=<that artist's ID>`. Picking from the list guarantees that the ID always exists.

### 6. Rendering pages and errors

`RenderPage` parses the requested template from `web/templates/`, fills it with data, and sends it. The page is first built in memory, so if the template fails, the user never receives a half-written page. Whenever a request fails validation or an internal step fails, `RenderError` renders `error.html` with the status code and a short message, using the same stylesheet as the rest of the site. A link returns the user to the home page.
