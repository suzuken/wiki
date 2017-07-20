package controller

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/csrf"
	"github.com/suzuken/wiki/httputil"
	"github.com/suzuken/wiki/model"
	"github.com/suzuken/wiki/view"
)

// Article is controller for requests to articles.
type Article struct {
	DB *sql.DB
}

// Root indicates / path as top page.
func (t *Article) Root(w http.ResponseWriter, r *http.Request) error {
	articles, err := model.ArticlesAll(t.DB)
	if err != nil {
		return err
	}
	return view.HTML(w, http.StatusOK, "index.tmpl", map[string]interface{}{
		"title":    "wiki wiki",
		"articles": articles,
		"request":  r,
	})
}

// Get returns specified article.
func (t *Article) Get(w http.ResponseWriter, r *http.Request) error {
	var id int64
	if _, err := fmt.Sscanf(r.URL.Path, "/articles/%d", id); err != nil {
		return err
	}
	article, err := model.ArticleOne(t.DB, id)
	if err != nil {
		return err
	}
	return view.HTML(w, http.StatusOK, "article.tmpl", map[string]interface{}{
		"title":   fmt.Sprintf("%s - go-wiki", article.Title),
		"article": article,
		"request": r,
	})
}

// Edit indicates edit page for certain article.
func (t *Article) Edit(w http.ResponseWriter, r *http.Request) error {
	var id int64
	if _, err := fmt.Sscanf(r.URL.Path, "/article/edit/%d", &id); err != nil {
		log.Printf("article.edit: %s", err)
		return err
	}
	article, err := model.ArticleOne(t.DB, id)
	if err != nil {
		return err
	}
	return view.HTML(w, http.StatusOK, "edit.tmpl", map[string]interface{}{
		"title":          fmt.Sprintf("%s - go-wiki", article.Title),
		"article":        article,
		"request":        r,
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}

// New works as endpoint to create new article.
// If successed, redirect to created one.
func (t *Article) New(w http.ResponseWriter, r *http.Request, m *model.Article) error {
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
		return err
	}
	http.Redirect(w, r, fmt.Sprintf("/article/%d", id), 301)
	return nil
}

// Update works for updating the specified article.
// After updating, redirect to one.
func (t *Article) Update(w http.ResponseWriter, r *http.Request, m *model.Article) error {
	if err := TXHandler(t.DB, func(tx *sql.Tx) error {
		if _, err := m.Update(tx); err != nil {
			return err
		}
		return tx.Commit()
	}); err != nil {
		return err
	}
	http.Redirect(w, r, fmt.Sprintf("/article/%d", m.ID), 301)
	return nil
}

// Save is endpoint for updating or creating documents.
// This accepts form request from browser.
// If id is specified, dealing with Update.
func (t *Article) Save(w http.ResponseWriter, r *http.Request) error {
	var article model.Article
	article.Body = r.PostFormValue("body")
	article.Title = r.PostFormValue("title")

	id := r.PostFormValue("id")
	if id == "" {
		return t.New(w, r, &article)
	}

	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}
	article.ID = aid
	return t.Update(w, r, &article)
}

// Delete is endpont for deleting the document.
func (t *Article) Delete(w http.ResponseWriter, r *http.Request) error {
	var article model.Article
	id := r.PostFormValue("id")
	if id == "" {
		return &httputil.HTTPError{Status: http.StatusBadRequest}
	}
	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}
	article.ID = aid
	if err := TXHandler(t.DB, func(tx *sql.Tx) error {
		if _, err := article.Delete(tx); err != nil {
			return err
		}
		return tx.Commit()
	}); err != nil {
		return err
	}

	http.Redirect(w, r, "/", 301)
	return nil
}
