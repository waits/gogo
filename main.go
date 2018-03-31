// Command gogo is a web server for hosting multiplayer Go games.
package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"

	"github.com/waits/gogo/handler"
	"github.com/waits/gogo/model"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/net/websocket"
)

var (
	host    = flag.String("host", "", "server hostname")
	dir     = flag.String("certs", "", "directory to store certificates in")
	devMode = flag.Bool("dev", false, "run in development mode")
)

func main() {
	flag.Parse()
	t := handler.LoadTemplates("template/")
	if *devMode {
		t = nil
	}
	env := &handler.Env{Templates: t, TemplatePath: "template/"}
	model.InitPool(0)

	http.Handle("/", handler.Handler{Env: env, Fn: handler.Static})
	http.Handle("/game/", handler.Handler{Env: env, Fn: handler.Game})
	http.Handle("/static/", http.FileServer(http.Dir("./")))
	http.Handle("/live/game/", websocket.Handler(handler.Live))

	if *devMode {
		log.Printf("Starting server at http://0.0.0.0:8080\n")
		log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
		return
	}

	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(*host),
		Cache:      autocert.DirCache(*dir),
	}

	s80 := &http.Server{
		Addr:    ":80",
		Handler: m.HTTPHandler(nil),
	}
	go s80.ListenAndServe()

	log.Printf("Starting server at https://" + *host)
	s := &http.Server{
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
	}
	log.Fatal(s.ListenAndServeTLS("", ""))
}
