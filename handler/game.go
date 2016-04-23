package handler

import (
	"encoding/json"
	"github.com/waits/gogo/model"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// GameHandler creates, updates, or loads a Game
func GameHandler(c *Context, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == "POST" {
		return createGame(c, w, r)
	}

	key := r.URL.Path[6:]
	game, err := model.Load(key)
	if err != nil {
		return http.StatusNotFound, err
	}

	if r.Method == "PUT" {
		return updateGame(r, game)
	} else if strings.Contains(key, "-vs-") {
		return http.StatusOK, RenderTemplate(c, w, "watch", game)
	} else {
		return http.StatusOK, RenderTemplate(c, w, "game", game)
	}
}

// LiveHandler sends game updates to a WebSocket connection
func LiveHandler(ws *websocket.Conn) {
	r := ws.Request()
	log.Printf("%s %s %s websocket", strings.Split(r.RemoteAddr, ":")[0], r.Method, r.URL.Path)

	key := r.URL.Path[11:]
	game, _ := model.Load(key)
	sendMsg := func(g *model.Game) {
		log.Printf("Sending WebSocket message for game %s", g.Key)
		err := json.NewEncoder(ws).Encode(g)
		if err != nil {
			log.Println(err.Error())
		}
	}
	sendMsg(game)
	model.Subscribe(game.Key, sendMsg)
}

func createGame(c *Context, w http.ResponseWriter, r *http.Request) (int, error) {
	size, _ := strconv.Atoi(r.FormValue("size"))
	handi, _ := strconv.Atoi(r.FormValue("handicap"))
	black := r.FormValue("black")
	white := r.FormValue("white")
	game, err := model.New(black, white, size, handi)
	if err != nil {
		return http.StatusBadRequest, err
	}
	http.Redirect(w, r, "/game/"+game.Key, 303)
	return http.StatusSeeOther, nil
}

func updateGame(r *http.Request, game *model.Game) (int, error) {
	color, _ := strconv.Atoi(r.FormValue("color"))
	var err error
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
