// Command playgo is a web server for hosting multiplayer Go games.
package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
)

var (
	httpAddr     = flag.String("http", "localhost:8080", "HTTP listen address")
	templatePath = flag.String("template", "template/", "path to template files")
	staticPath   = flag.String("static", "static/", "path to static files")
	reload       = flag.Bool("reload", false, "reload templates for every request")
)

type Context struct {
	Templates map[string]*template.Template
}

func main() {
	flag.Parse()
	t := loadTemplates()
	if *reload {
		t = nil
	}
	c := &Context{t}

	log.Printf("Starting server at http://%s\n", *httpAddr)
	http.Handle("/", reqHandler{c, rootHandler})
	http.Handle("/game/", reqHandler{c, gameHandler})
	http.ListenAndServe(*httpAddr, nil)
}
