package controller

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/suzuken/wiki/model"
)

// User is controller for requests to user.
type User struct {
	DB *sql.DB
}

// SignUp makes user signup.
func (u *User) SignUp(c context.Context) {
	var m model.User
	m.Name = c.PostForm("name")
	m.Email = c.PostForm("email")
	password := c.PostForm("password")

	b, err := model.UserExists(u.DB, m.Email)
	if err != nil {
		log.Printf("query error: %s", err)
		c.String(500, "db error")
		return
	}
	if b {
		c.String(200, "given email address is already used.")
		return
	}

	TXHandler(c, u.DB, func(tx *sql.Tx) error {
		if _, err := m.Insert(tx, password); err != nil {
			return err
		}
		return tx.Commit()
	})

	c.Redirect(301, "/")
}

// Login try login.
func (u *User) Login(c *gin.Context) {
	m, err := model.Auth(u.DB, c.PostForm("email"), c.PostForm("password"))
	if err != nil {
		log.Printf("auth failed: %s", err)
		c.String(500, "can't auth")
		return
	}

	log.Printf("authed: %#v", m)

	sess := sessions.Default(c)
	sess.Set("uid", m.ID)
	sess.Set("name", m.Name)
	sess.Set("email", m.Email)
	sess.Save()

	c.Redirect(301, "/")
}

// Logout makes user logged out.
func (u *User) Logout(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Options(sessions.Options{MaxAge: -1})
	sess.Clear()
	sess.Save()

	// clear cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:    "wikisession",
		Value:   "",
		Path:    "/",
		Expires: time.Now().AddDate(0, -1, 0),
	})

	c.Redirect(301, "/")
}

// LoggedIn returns if current session user is logged in or not.
func LoggedIn(r *http.Request) bool {
	if r == nil {
		return false
	}
	re := sessions.GetRegistry(r)
	re.Get("uid")
	return sess.Get("uid") != nil && sess.Get("name") != nil && sess.Get("email") != nil
}

// CurrentName returns current user name who logged in.
func CurrentName(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return sessions.Default(c).Get("name").(string)
}

// AuthRequired returns a handler function which checks
// if user logged in or not.
func AuthRequired() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !LoggedIn(r) {
			http.Error(w, "abort", http.StatusUnauthorized)
		}
	}
}
