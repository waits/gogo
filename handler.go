package main

import (
	"errors"
	"log"
	"net/http"
	"playgo/model"
	"strconv"
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
			http.Error(w, err.Error(), status)
		case http.StatusBadRequest:
			http.Error(w, err.Error(), status)
		default:
			status = http.StatusInternalServerError
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
	if r.Method == "POST" {
		size, _ := strconv.Atoi(r.FormValue("size"))
		game, err := model.New(r.FormValue("black"), r.FormValue("white"), size)
		if err != nil {
			return http.StatusBadRequest, err
		}
		http.Redirect(w, r, "/game/"+game.Id, 303)
		return http.StatusSeeOther, nil
	} else {
		id := r.URL.Path[6:]
		game, err := model.Load(id)
		if err != nil {
			return http.StatusNotFound, err
		}
		if r.Method == "PATCH" {
			x, _ := strconv.Atoi(r.FormValue("x"))
			y, _ := strconv.Atoi(r.FormValue("y"))
			err = game.Move(x, y)
			if err != nil {
				return http.StatusBadRequest, err
			}
			return http.StatusOK, nil
		} else {
			return http.StatusOK, renderTemplate(c, w, "game", game)
		}
	}
}
