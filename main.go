// Command playgo is a web server for hosting multiplayer Go games.
package main

import (
	"flag"
	"gogo/handler"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

var (
	httpAddr     = flag.String("http", "localhost:8080", "HTTP listen address")
	templatePath = flag.String("template", "template/", "path to template files")
	staticPath   = flag.String("static", "static/", "path to static files")
	reload       = flag.Bool("reload", false, "reload templates for every request")
)

func main() {
	flag.Parse()
	t := handler.LoadTemplates(*templatePath)
	if *reload {
		t = nil
	}
	c := &handler.Context{t, *templatePath}

	log.Printf("Starting server at http://%s\n", *httpAddr)
	http.Handle("/", handler.Handler{c, handler.StaticHandler})
	http.Handle("/game/", handler.Handler{c, handler.GameHandler})
	http.Handle("/static/", http.FileServer(http.Dir("./")))
	http.Handle("/live/game/", websocket.Handler(handler.LiveHandler))
	http.ListenAndServe(*httpAddr, nil)
}
