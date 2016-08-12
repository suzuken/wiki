package controller

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/suzuken/wiki/model"
)

// User is controller for requests to user.
type User struct {
	DB *sql.DB
}

func (u *User) SignUp(c *gin.Context) {
	var m model.User
	m.Name = c.PostForm("name")
	m.Email = c.PostForm("email")
	password := c.PostForm("password")

	TXHandler(c, u.DB, func(tx *sql.Tx) error {
		if _, err := m.Insert(tx, password); err != nil {
			return err
		}
		return tx.Commit()
	})

	// TODO should be login state here
	c.Redirect(301, "/")
}

func (u *User) Login(c *gin.Context) {
	m, err := model.Auth(u.DB, c.PostForm("email"), c.PostForm("password"))
	if err != nil {
		log.Printf("auth failed: %s", err)
		c.String(500, "can't auth")
		return
	}

	log.Printf("authed: %v", m)

	// TODO should be login state here
	c.Redirect(301, "/")
}
