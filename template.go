package main

import (
	"errors"
	"html/template"
	"net/http"
	"path/filepath"
)

// Loads all templates into memory (layout and content)
func loadTemplates() map[string]*template.Template {
	templates := make(map[string]*template.Template)
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
	return templates
}

// Renders an HTML template from the cache using provided data
func renderTemplate(c *Context, w http.ResponseWriter, name string, data interface{}) error {
	tmpls := c.Templates
	if c.Templates == nil {
		tmpls = loadTemplates()
	}
	tmpl, ok := tmpls[name+".tmpl"]
	if !ok {
		return errors.New("renderTemplate: template does not exist")
	}
	err := tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		return err
	}
	return nil
}
