// Command playgo is a web server for hosting multiplayer Go games.
package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

const host = "localhost"
const port = "8080"

type Game struct {
	Id    int
	White string
	Black string
}

// Wraps a route handler in a closure, then logs the request address, method,
// and path, plus the status code returned by the handler
func makeHandler(fn func(http.ResponseWriter, *http.Request) int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := fn(w, r)
		log.Printf("%s %s %s %d", strings.Split(r.RemoteAddr, ":")[0], r.Method, r.URL.Path, status)
	}
}

// Renders the home template
func rootHandler(w http.ResponseWriter, r *http.Request) int {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return 404
	}
	err := renderTemplate(w, "index", nil)
	if err != nil {
		return 500
	}
	return 200
}

// Renders the game template
func gameHandler(w http.ResponseWriter, r *http.Request) int {
	id, err := strconv.Atoi(r.URL.Path[6:])
	if err != nil {
		http.NotFound(w, r)
		return 404
	}
	game := loadGame(id)
	err = renderTemplate(w, "game", game)
	if err != nil {
		return 500
	}
	return 200
}

// Loads a game for a provided id
func loadGame(id int) *Game {
	return &Game{Id: id, White: "John", Black: "Frank"}
}

func main() {
	addr := host + ":" + port
	log.Printf("Starting server at http://%s\n", addr)
	http.HandleFunc("/", makeHandler(rootHandler))
	http.HandleFunc("/game/", makeHandler(gameHandler))
	http.ListenAndServe(addr, nil)
}
