//go:generate scaneo $GOFILE

package model

import (
	"database/sql"
	"time"
)

type Article struct {
	ID      int64      `json:"id"`
	Title   string     `json:"title"`
	Body    string     `json:"body"`
	Created *time.Time `json:"created"`
	Updated *time.Time `json:"updated"`
}

func ArticlesAll(db *sql.DB) ([]Article, error) {
	rows, err := db.Query(`select * from articles`)
	if err != nil {
		return nil, err
	}
	return ScanArticles(rows)
}

func ArticleOne(db *sql.DB, id int64) (Article, error) {
	return ScanArticle(db.QueryRow(`select * from articles where article_id = ?`, id))
}

func (t *Article) Update(tx *sql.Tx) (sql.Result, error) {
	stmt, err := tx.Prepare(`
	update articles
		set title = ?, body = ?
		where article_id = ?
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	return stmt.Exec(t.Title, t.Body, t.ID)
}

func (t *Article) Insert(tx *sql.Tx) (sql.Result, error) {
	stmt, err := tx.Prepare(`
	insert into articles (title, body)
	values(?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	return stmt.Exec(t.Title, t.Body)
}

func (t *Article) Delete(tx *sql.Tx) (sql.Result, error) {
	stmt, err := tx.Prepare(`delete from articles where article_id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	return stmt.Exec(t.ID)
}

// ArticlesDeleteAllはすべてのタスクを消去します。
// テストのために使用されます。
func ArticlesDeleteAll(tx *sql.Tx) (sql.Result, error) {
	return tx.Exec(`truncate table articles`)
}
