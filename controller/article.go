package controller

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/csrf"
	"github.com/julienschmidt/httprouter"
	"github.com/suzuken/wiki/model"
	"github.com/suzuken/wiki/view"
)

// Article is controller for requests to articles.
type Article struct {
	DB *sql.DB
}

// Root indicates / path as top page.
func (t *Article) Root(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	articles, err := model.ArticlesAll(t.DB)
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, fmt.Sprintf("%s", err))
		return
	}
	view.HTML(w, http.StatusOK, "index.tmpl", map[string]interface{}{
		"title":    "wiki wiki",
		"articles": articles,
		"request":  r,
	})
}

// Get returns specified article.
func (t *Article) Get(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id := params.ByName("id")
	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		Error(w, err, 500)
		return
	}
	article, err := model.ArticleOne(t.DB, aid)
	if err != nil {
		Error(w, err, 500)
		return
	}
	view.HTML(w, http.StatusOK, "article.tmpl", map[string]interface{}{
		"title":   fmt.Sprintf("%s - go-wiki", article.Title),
		"article": article,
		"request": r,
	})
}

// Edit indicates edit page for certain article.
func (t *Article) Edit(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id := params.ByName("id")
	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, fmt.Sprintf("%s", err))
		return
	}
	article, err := model.ArticleOne(t.DB, aid)
	if err != nil {
		Error(w, err, 500)
		return
	}
	view.HTML(w, http.StatusOK, "edit.tmpl", map[string]interface{}{
		"title":          fmt.Sprintf("%s - go-wiki", article.Title),
		"article":        article,
		"request":        r,
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}

// New works as endpoint to create new article.
// If successed, redirect to created one.
func (t *Article) New(w http.ResponseWriter, r *http.Request, m *model.Article) {
	var id int64
	if err := TXHandler(t.DB, func(tx *sql.Tx) error {
		result, err := m.Insert(tx)
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		id, err = result.LastInsertId()
		return err
	}); err != nil {
		Error(w, err, 500)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/article/%d", id), 301)
}

// Update works for updating the specified article.
// After updating, redirect to one.
func (t *Article) Update(w http.ResponseWriter, r *http.Request, m *model.Article) {
	if err := TXHandler(t.DB, func(tx *sql.Tx) error {
		if _, err := m.Update(tx); err != nil {
			return err
		}
		return tx.Commit()
	}); err != nil {
		Error(w, err, 500)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/article/%d", m.ID), 301)
}

// Save is endpoint for updating or creating documents.
// This accepts form request from browser.
// If id is specified, dealing with Update.
func (t *Article) Save(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var article model.Article
	article.Body = r.PostFormValue("body")
	article.Title = r.PostFormValue("title")

	id := r.PostFormValue("id")
	if id == "" {
		t.New(w, r, &article)
		return
	}

	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		Error(w, err, 500)
		return
	}
	article.ID = aid
	t.Update(w, r, &article)
}

// Delete is endpont for deleting the document.
func (t *Article) Delete(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var article model.Article
	id := r.PostFormValue("id")
	if id == "" {
		http.Error(w, "id required but not specified", http.StatusBadRequest)
		return
	}
	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		Error(w, err, 500)
		return
	}
	article.ID = aid
	if err := TXHandler(t.DB, func(tx *sql.Tx) error {
		if _, err := article.Delete(tx); err != nil {
			return err
		}
		return tx.Commit()
	}); err != nil {
		Error(w, err, 500)
		return
	}

	http.Redirect(w, r, "/", 301)
}
