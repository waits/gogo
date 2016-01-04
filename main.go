// Command playgo is a web server for hosting multiplayer Go games.
package main

import (
	"flag"
	"log"
	"net/http"
)

var (
	httpAddr     = flag.String("http", "localhost:8080", "HTTP listen address")
	templatePath = flag.String("template", "template/", "path to template files")
	staticPath   = flag.String("static", "static/", "path to static files")
	reload       = flag.Bool("reload", false, "reload templates on every page load")
)

type Game struct {
	Id    int
	White string
	Black string
}

func init() {
	flag.Parse()
	loadTemplates()
}

func main() {
	log.Printf("Starting server at http://%s\n", *httpAddr)
	http.Handle("/", reqHandler(rootHandler))
	http.Handle("/game/", reqHandler(gameHandler))
	http.ListenAndServe(*httpAddr, nil)
}

// Loads a game for a provided id
func loadGame(id int) *Game {
	return &Game{Id: id, White: "John", Black: "Frank"}
}
