package main

import (
	"errors"
	"log"
	"net/http"
	"playgo/model"
	"strings"
)

type reqHandler struct {
	*Context
	Fn func(*Context, http.ResponseWriter, *http.Request) (int, error)
}

// ServeHTTP is called on a reqHandler by net/http; Satisfies http.Handler
func (h reqHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, err := h.Fn(h.Context, w, r)
	if err != nil {
		switch status {
		case http.StatusNotFound:
			http.NotFound(w, r)
		default:
			status = 500
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
	log.Printf("%s %s %s %d", strings.Split(r.RemoteAddr, ":")[0], r.Method, r.URL.Path, status)
}

// Renders the home and about templates
func rootHandler(c *Context, w http.ResponseWriter, r *http.Request) (int, error) {
	switch r.URL.Path {
	case "/":
		return http.StatusOK, renderTemplate(c, w, "home", nil)
	case "/about":
		return http.StatusOK, renderTemplate(c, w, "about", nil)
	default:
		return http.StatusNotFound, errors.New("handler: page not found")
	}
}

// Renders the game template
func gameHandler(c *Context, w http.ResponseWriter, r *http.Request) (int, error) {
	id := r.URL.Path[6:]
	/*
		if err != nil {
			return http.StatusNotFound, err
		}
	*/
	var game *model.Game
	if id == "new" {
		game = model.NewGame("Bob", "Mary")
		http.Redirect(w, r, "/game/"+game.Id, 303)
		return 303, nil
	} else {
		game = model.LoadGame(id)
	}
	return http.StatusOK, renderTemplate(c, w, "game", game)
}
