package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type reqHandler func(http.ResponseWriter, *http.Request) (int, error)

// ServeHTTP is called on a reqHandler by net/http; Satisfies http.Handler
func (fn reqHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, err := fn(w, r)
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
func rootHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	switch r.URL.Path {
	case "/":
		return http.StatusOK, renderTemplate(w, "home", nil)
	case "/about":
		return http.StatusOK, renderTemplate(w, "about", nil)
	default:
		return http.StatusNotFound, errors.New("handler: page not found")
	}
}

// Renders the game template
func gameHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.URL.Path[6:])
	if err != nil {
		return http.StatusNotFound, err
	}
	game := loadGame(id)
	return http.StatusOK, renderTemplate(w, "game", game)
}
