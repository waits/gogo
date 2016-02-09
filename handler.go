package main

import (
	"encoding/json"
	"errors"
	"gogo/model"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
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
			http.NotFound(w, r)
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
		games := model.Recent(20)
		return http.StatusOK, renderTemplate(c, w, "home", games)
	case "/new":
		return http.StatusOK, renderTemplate(c, w, "new", nil)
	default:
		return http.StatusNotFound, errors.New("handler: page not found")
	}
}

// Renders the game template
func gameHandler(c *Context, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == "POST" {
		size, _ := strconv.Atoi(r.FormValue("size"))
		black := r.FormValue("black")
		white := r.FormValue("white")
		game, err := model.New(black, white, size)
		if err != nil {
			return http.StatusBadRequest, err
		}
		http.Redirect(w, r, "/game/"+game.Key, 303)
		return http.StatusSeeOther, nil
	}

	key := r.URL.Path[6:]
	game, err := model.Load(key)
	if err != nil {
		return http.StatusNotFound, err
	}
	if r.Method == "PATCH" {
		color, _ := strconv.Atoi(r.FormValue("color"))
		if pass := r.FormValue("pass"); pass != "" {
			err = game.Pass(color)
		} else {
			x, _ := strconv.Atoi(r.FormValue("x"))
			y, _ := strconv.Atoi(r.FormValue("y"))
			err = game.Move(color, x, y)
		}
		if err != nil {
			return http.StatusBadRequest, err
		}
		return http.StatusOK, nil
	}

	return http.StatusOK, renderTemplate(c, w, "game", game)
}

func watchHandler(c *Context, w http.ResponseWriter, r *http.Request) (int, error) {
	key := r.URL.Path[7:]
	game, err := model.Load(key)
	if err != nil {
		return http.StatusNotFound, err
	}
	return http.StatusOK, renderTemplate(c, w, "watch", game)
}

// Sends game updates to a WebSocket connection
func liveHandler(ws *websocket.Conn) {
	r := ws.Request()
	log.Printf("%s %s %s websocket", strings.Split(r.RemoteAddr, ":")[0], r.Method, r.URL.Path)

	key := r.URL.Path[11:]
	game, _ := model.Load(key)
	sendMsg := func(g *model.Game) {
		log.Printf("Sending WebSocket message for game %s", g.Key)
		err := json.NewEncoder(ws).Encode(g)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}
	sendMsg(game)
	model.Subscribe(game.Key, sendMsg)
}
