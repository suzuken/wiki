package view

import (
	"html/template"
	"io"
	"net/http"

	"github.com/gorilla/csrf"
)

var executor TemplateExecutor

func Init(funcs template.FuncMap, debug bool) {
	if debug {
		executor = DebugTemplateExecutor{
			Glob:  "templates/*",
			Funcs: funcs,
		}
		return
	}

	executor = CachedTemplateExecutor{
		Template: template.Must(template.New("").Funcs(funcs).ParseGlob("templates/*")),
	}
}

type TemplateExecutor interface {
	ExecuteTemplate(w io.Writer, name string, data interface{}) error
}

type DebugTemplateExecutor struct {
	Glob  string
	Funcs template.FuncMap
}

func (e DebugTemplateExecutor) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	return template.Must(template.New("").Funcs(e.Funcs).ParseGlob(e.Glob)).ExecuteTemplate(w, name, data)
}

type CachedTemplateExecutor struct {
	Template *template.Template
}

func (e CachedTemplateExecutor) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	return e.Template.ExecuteTemplate(w, name, data)
}

// HTML render view
func HTML(w http.ResponseWriter, status int, name string, data map[string]interface{}) error {
	w.WriteHeader(status)
	return executor.ExecuteTemplate(w, name, data)
}

// Default is shorthands for rendering template.
// This includes HTTP response writer and HTTP request object for calling helper funcs.
func Default(w http.ResponseWriter, r *http.Request, status int, name string, data map[string]interface{}) error {
	m := map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
		"request":        r,
		"writer":         w,
	}
	if len(data) > 0 {
		for k, v := range data {
			m[k] = v
		}
	}
	return HTML(w, status, name, m)
}
