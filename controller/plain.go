package controller

import (
	"io"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/suzuken/wiki/view"
)

func AuthTestHandler(w http.ResponseWriter, _ *http.Request) error {
	w.WriteHeader(200)
	_, err := io.WriteString(w, "your're authed")
	return err
}

func NewArticleHandler(w http.ResponseWriter, r *http.Request) error {
	return view.HTML(w, 200, "new.tmpl", map[string]interface{}{
		"title":          "New: go-wiki",
		csrf.TemplateTag: csrf.TemplateField(r),
		"request":        r,
	})
}
