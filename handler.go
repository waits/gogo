package main

import "net/http"
import "strconv"

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
