package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/suzuken/wiki/model"
)

// Article is controller for requests to articles.
type Article struct {
	DB *sql.DB
}

// Root indicates / path as top page.
func (t *Article) Root(c echo.Context) error {
	articles, err := model.ArticlesAll(t.DB)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "index.tmpl", echo.Map{
		"title":    "wiki wiki",
		"articles": articles,
		"context":  c,
	})
}

// Get returns specified article.
func (t *Article) Get(c echo.Context) error {
	id := c.Param("id")
	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}
	article, err := model.ArticleOne(t.DB, aid)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "article.tmpl", echo.Map{
		"title":   fmt.Sprintf("%s - go-wiki", article.Title),
		"article": article,
		"context": c,
	})
}

// Edit indicates edit page for certain article.
func (t *Article) Edit(c echo.Context) error {
	id := c.Param("id")
	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}
	article, err := model.ArticleOne(t.DB, aid)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "edit.tmpl", echo.Map{
		"title":   fmt.Sprintf("%s - go-wiki", article.Title),
		"article": article,
		"context": c,

		// TODO should be wrap this work around
		"csrf": c.Get(middleware.DefaultCSRFConfig.ContextKey).(string),
	})
}

// New works as endpoint to create new article.
// If successed, redirect to created one.
func (t *Article) New(c echo.Context, m *model.Article) error {
	var id int64
	TXHandler(c, t.DB, func(tx *sql.Tx) error {
		result, err := m.Insert(tx)
		if err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		id, err = result.LastInsertId()
		return err
	})
	return c.Redirect(301, fmt.Sprintf("/article/%d", id))
}

// Update works for updating the specified article.
// After updating, redirect to one.
func (t *Article) Update(c echo.Context, m *model.Article) error {
	TXHandler(c, t.DB, func(tx *sql.Tx) error {
		if _, err := m.Update(tx); err != nil {
			return err
		}
		return tx.Commit()
	})
	return c.Redirect(301, fmt.Sprintf("/article/%d", m.ID))
}

// Save is endpoint for updating or creating documents.
// This accepts form request from browser.
// If id is specified, dealing with Update.
func (t *Article) Save(c echo.Context) error {
	var article model.Article
	article.Body = c.FormValue("body")
	article.Title = c.FormValue("title")

	id := c.FormValue("id")
	if id == "" {
		return t.New(c, &article)
	}

	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}
	article.ID = aid
	return t.Update(c, &article)
}

// Delete is endpont for deleting the document.
func (t *Article) Delete(c echo.Context) error {
	var article model.Article
	id := c.FormValue("id")
	if id == "" {
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}
	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}
	article.ID = aid
	TXHandler(c, t.DB, func(tx *sql.Tx) error {
		if _, err := article.Delete(tx); err != nil {
			return err
		}
		return tx.Commit()
	})

	return c.Redirect(301, "/")
}
