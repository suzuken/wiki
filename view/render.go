package view

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gorilla/csrf"
	"github.com/mattn/go-slim"
)

var executor TemplateExecutor

func Init(funcs slim.Funcs) {
	matches, err := filepath.Glob("/Users/suzuken/src/github.com/suzuken/wiki/slimtemplates/*.tmpl")
	if err != nil {
		panic(err)
	}
	fmt.Printf("matches = %+v\n", matches)
	templates := make(map[string]*slim.Template)

	for _, m := range matches {
		tmpl, err := slim.ParseFile(m)
		if err != nil {
			panic(err)
		}
		tmpl.FuncMap(funcs)
		templates[m] = tmpl
	}
	fmt.Printf("templates = %+v\n", templates)
	executor = CachedTemplateExecutor{
		Template: templates,
	}
}

type TemplateExecutor interface {
	Execute(w io.Writer, name string, data interface{}) error
}

type CachedTemplateExecutor struct {
	Template map[string]*slim.Template
}

func (e CachedTemplateExecutor) Execute(w io.Writer, name string, data interface{}) error {
	tmpl, ok := e.Template[name]
	if !ok {
		return errors.New("template not found")
	}
	return tmpl.Execute(w, data)
}

func Execute(w io.Writer, name string, data map[string]interface{}) error {
	return executor.Execute(w, name, data)
}

// HTML render view
func HTML(w http.ResponseWriter, status int, name string, data map[string]interface{}) error {
	w.WriteHeader(status)
	return executor.Execute(w, name, data)
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
