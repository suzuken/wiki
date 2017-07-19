package view

import (
	"html/template"
	"net/http"
)

// templates is singleton.
var templates = template.New("")

func Init() {
	templates = template.Must(templates.ParseGlob("templates/*"))
}

func Funcs(m template.FuncMap) {
	templates.Funcs(m)
}

func HTML(w http.ResponseWriter, status int, name string, data map[string]interface{}) error {
	w.WriteHeader(status)
	return templates.ExecuteTemplate(w, name, data)
}
