package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/suzuken/wiki/model"
	csrf "github.com/utrack/gin-csrf"

	"github.com/gin-gonic/gin"
)

var comments map[string][]string

func init() {
	comments = make(map[string][]string, 0)
}

// Article is controller for requests to articles.
type Article struct {
	DB *sql.DB
}

// Root indicates / path as top page.
func (t *Article) Root(c *gin.Context) {
	articles, err := model.ArticlesAll(t.DB)
	if err != nil {
		c.String(500, "%s", err)
		return
	}
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":    "wiki wiki",
		"articles": articles,
		"context":  c,
	})
}

// Get returns specified article.
func (t *Article) Get(c *gin.Context) {
	id := c.Param("id")
	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.String(500, "%s", err)
		return
	}
	article, err := model.ArticleOne(t.DB, aid)
	if err != nil {
		c.String(500, "%s", err)
		return
	}
	c.HTML(http.StatusOK, "article.tmpl", gin.H{
		"title":    fmt.Sprintf("%s - go-wiki", article.Title),
		"article":  article,
		"csrf":     csrf.GetToken(c),
		"context":  c,
		"comments": comments[id],
	})
}

// Edit indicates edit page for certain article.
func (t *Article) Edit(c *gin.Context) {
	id := c.Param("id")
	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.String(500, "%s", err)
		return
	}
	article, err := model.ArticleOne(t.DB, aid)
	if err != nil {
		c.String(500, "%s", err)
		return
	}
	c.HTML(http.StatusOK, "edit.tmpl", gin.H{
		"title":   fmt.Sprintf("%s - go-wiki", article.Title),
		"article": article,
		"context": c,
		"csrf":    csrf.GetToken(c),
	})
}

// New works as endpoint to create new article.
// If successed, redirect to created one.
func (t *Article) New(c *gin.Context, m *model.Article) {
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
	c.Redirect(301, fmt.Sprintf("/article/%d", id))
}

// Update works for updating the specified article.
// After updating, redirect to one.
func (t *Article) Update(c *gin.Context, m *model.Article) {
	TXHandler(c, t.DB, func(tx *sql.Tx) error {
		if _, err := m.Update(tx); err != nil {
			return err
		}
		return tx.Commit()
	})
	c.Redirect(301, fmt.Sprintf("/article/%d", m.ID))
}

func (t *Article) Comment(c *gin.Context) {
	body := c.PostForm("body")
	id := c.PostForm("id")
	comments[id] = append(comments[id], body)
	c.Redirect(301, fmt.Sprintf("/article/%s", id))
}

// Save is endpoint for updating or creating documents.
// This accepts form request from browser.
// If id is specified, dealing with Update.
func (t *Article) Save(c *gin.Context) {
	var article model.Article
	article.Body = c.PostForm("body")
	article.Title = c.PostForm("title")

	id := c.PostForm("id")
	if id == "" {
		t.New(c, &article)
		return
	}

	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.String(500, "%s", err)
		return
	}
	article.ID = aid
	t.Update(c, &article)
}

// Delete is endpont for deleting the document.
func (t *Article) Delete(c *gin.Context) {
	var article model.Article
	id := c.PostForm("id")
	if id == "" {
		c.Abort()
		return
	}
	aid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.String(500, "%s", err)
		return
	}
	article.ID = aid
	TXHandler(c, t.DB, func(tx *sql.Tx) error {
		if _, err := article.Delete(tx); err != nil {
			return err
		}
		return tx.Commit()
	})

	c.Redirect(301, "/")
}
