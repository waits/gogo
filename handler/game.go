package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/waits/gogo/model"
	"golang.org/x/net/websocket"
)

// Game creates, updates, or loads a game.
func Game(env *Env, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == "POST" {
		return createGame(env, w, r)
	}

	key := r.URL.Path[6:]
	game, err := model.Load(key)
	if err != nil {
		return http.StatusNotFound, err
	}

	cookie, _ := r.Cookie(key)

	if r.Method == "PUT" {
		if cookie == nil {
			return joinGame(w, r, game)
		}
		return updateGame(w, r, game, cookie.Value)
	} else if cookie != nil {
		return http.StatusOK, RenderTemplate(env, w, "game", game)
	} else if game.Black == "" || game.White == "" {
		return http.StatusOK, RenderTemplate(env, w, "join", game)
	}

	return http.StatusOK, RenderTemplate(env, w, "watch", game)
}

// Live sends game updates to a WebSocket connection.
func Live(ws *websocket.Conn) {
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

func createGame(env *Env, w http.ResponseWriter, r *http.Request) (int, error) {
	size, _ := strconv.Atoi(r.FormValue("size"))
	handi, _ := strconv.Atoi(r.FormValue("handicap"))
	name := r.FormValue("name")
	color := r.FormValue("color")

	game, err := model.New(name, color, size, handi)
	if err != nil {
		return http.StatusBadRequest, err
	}

	http.SetCookie(w, &http.Cookie{Name: game.Key, Value: color, MaxAge: 604800})
	http.Redirect(w, r, "/game/"+game.Key, 303)
	return http.StatusSeeOther, nil
}

func joinGame(w http.ResponseWriter, r *http.Request, game *model.Game) (int, error) {
	name := r.FormValue("name")
	color := game.Join(name)
	http.SetCookie(w, &http.Cookie{Name: game.Key, Value: color, MaxAge: 604800})
	http.Redirect(w, r, "/game/"+game.Key, 303)
	return http.StatusSeeOther, nil
}

func updateGame(w http.ResponseWriter, r *http.Request, game *model.Game, color string) (int, error) {
	var err error

	if pass := r.FormValue("pass"); pass != "" {
		err = game.Pass(color)
	} else {
		x, _ := strconv.Atoi(r.FormValue("x"))
		y, _ := strconv.Atoi(r.FormValue("y"))
		p := model.Point{X: x, Y: y}
		err = game.Move(color, p)
	}

	if err != nil {
		return http.StatusBadRequest, err
	}
	return http.StatusOK, nil
}
