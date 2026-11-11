# Groupie Tracker

A Go web app that shows bands and artists from the [Groupie Trackers API](https://groupietrackers.herokuapp.com/api). It uses only the Go standard library and no JavaScript.

## Features

- Gallery of all artists
- Artist pages with members, creation year, first album and concert dates
- Search by artist, member, location, first album date or creation year
- Random artist button
- Styled error pages (400, 404, 405, 500)

## Usage

```bash
go run .
```

Then open http://localhost:8080. You need an internet connection because the data comes from the API.

## Routes

| Route | Description |
|---|---|
| `/` | All artists |
| `/artist?id=<id>` | One artist |
| `/search?q=<text>` | Search results |
| `/random` | Redirects to a random artist |

## Testing

```bash
go test ./... -v
```

## Authors

- Yaser — search and random artist
- Hussain — HTML templates and CSS
- Sayed Hashim — data, API, routing and errors
