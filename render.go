package main

import "html/template"
import "net/http"

var templates = template.Must(template.ParseFiles(pathTo("game"), pathTo("index")))

// Renders an HTML template using provided data
func renderTemplate(w http.ResponseWriter, t string, data interface{}) error {
	err := templates.ExecuteTemplate(w, t+".tmpl", data)
	if err != nil {
		return err
	}
	return nil
}

// Returns the relative path to a template
func pathTo(t string) string {
	return "template/" + t + ".tmpl"
}
