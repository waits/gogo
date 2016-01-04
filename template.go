package main

import (
	"errors"
	"html/template"
	"net/http"
	"path/filepath"
)

var templates map[string]*template.Template

// Loads all templates into memory (layout and content)
func loadTemplates() {
	templates = make(map[string]*template.Template)
	pages, err := filepath.Glob(*templatePath + "content/*.tmpl")
	if err != nil {
		panic(err)
	}
	layouts, err := filepath.Glob(*templatePath + "*.tmpl")
	if err != nil {
		panic(err)
	}
	for _, page := range pages {
		files := append(layouts, page)
		templates[filepath.Base(page)] = template.Must(template.ParseFiles(files...))
	}
}

// Renders an HTML template from the cache using provided data
func renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	if *reload {
		loadTemplates()
	}
	tmpl, ok := templates[name+".tmpl"]
	if !ok {
		return errors.New("renderTemplate: template does not exist")
	}
	err := tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		return err
	}
	return nil
}
