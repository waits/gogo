package handler

import (
	"errors"
	"github.com/waits/gogo/model"
	"html/template"
	"log"
	"net/http"
	"strings"
)

// Context holds a map of cached templates
type Context struct {
	Templates    map[string]*template.Template
	TemplatePath string
}

// Handler wraps a route handler with a Context
type Handler struct {
	*Context
	Fn func(*Context, http.ResponseWriter, *http.Request) (int, error)
}

// ServeHTTP is called on a reqHandler by net/http; Satisfies http.Handler
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m := r.FormValue("_method"); len(m) > 0 {
		r.Method = strings.ToUpper(m)
	}
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

// StaticHandler responds to static routes not covered by another handler
func StaticHandler(c *Context, w http.ResponseWriter, r *http.Request) (int, error) {
	switch r.URL.Path {
	case "/":
		games := model.Recent()
		return http.StatusOK, RenderTemplate(c, w, "home", games)
	case "/new":
		return http.StatusOK, RenderTemplate(c, w, "new", nil)
	case "/help":
		return http.StatusOK, RenderTemplate(c, w, "help", nil)
	default:
		return http.StatusNotFound, errors.New("handler: page not found")
	}
}
