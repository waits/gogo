package main

import "net/http"
import "strconv"

// Renders the home and about templates
func rootHandler(w http.ResponseWriter, r *http.Request) int {
	var err error
	switch r.URL.Path {
	case "/":
		err = renderTemplate(w, "home", nil)
	case "/about":
		err = renderTemplate(w, "about", nil)
	default:
		http.NotFound(w, r)
		return 404
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
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
		http.Error(w, err.Error(), 500)
		return 500
	}
	return 200
}
