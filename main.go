// Command gogo is a web server for hosting multiplayer Go games.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/waits/gogo/handler"
	"github.com/waits/gogo/model"
	"golang.org/x/net/websocket"
)

var (
	db       = flag.Int("db", 0, "Redis database number")
	addr     = flag.String("addr", "localhost:8080", "Address to listen on")
	reload   = flag.Bool("reload", false, "reload templates for every request")
	certFile = flag.String("cert", "", "path to certificate chain")
	keyFile  = flag.String("key", "", "path to private key")
)

func main() {
	flag.Parse()
	t := handler.LoadTemplates("template/")
	if *reload {
		t = nil
	}
	env := &handler.Env{Templates: t, TemplatePath: "template/"}
	model.InitPool(*db)

	log.Printf("Starting server at http://%s\n", *addr)
	http.Handle("/", handler.Handler{Env: env, Fn: handler.StaticHandler})
	http.Handle("/game/", handler.Handler{Env: env, Fn: handler.GameHandler})
	http.Handle("/static/", http.FileServer(http.Dir("./")))
	http.Handle("/live/game/", websocket.Handler(handler.LiveHandler))

	if *certFile != "" && *keyFile != "" {
		log.Fatal(http.ListenAndServeTLS(*addr, *certFile, *keyFile, nil))
	}
	log.Fatal(http.ListenAndServe(*addr, nil))
}
