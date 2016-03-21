package handler

import (
	"errors"
	"html/template"
	"net/http"
	"path/filepath"
)

// LoadTemplates loads all templates into memory (layout and content)
func LoadTemplates(path string) map[string]*template.Template {
	templates := make(map[string]*template.Template)
	pages, err := filepath.Glob(path + "content/*.tmpl")
	if err != nil {
		panic(err)
	}
	layouts, err := filepath.Glob(path + "*.tmpl")
	if err != nil {
		panic(err)
	}
	// 	funcMap := template.FuncMap{"incr": incr}
	for _, page := range pages {
		files := append(layouts, page)
		templates[filepath.Base(page)] = template.Must(template.New(page).ParseFiles(files...))
	}
	return templates
}

// RenderTemplate renders an HTML template from the cache using provided data
func RenderTemplate(c *Context, w http.ResponseWriter, name string, data interface{}) error {
	tmpls := c.Templates
	if c.Templates == nil {
		tmpls = LoadTemplates(c.TemplatePath)
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
