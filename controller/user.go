package controller

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/ipfans/echo-session"
	"github.com/labstack/echo"
	"github.com/suzuken/wiki/model"
)

// User is controller for requests to user.
type User struct {
	DB *sql.DB
}

// SignUp makes user signup.
func (u *User) SignUp(c echo.Context) error {
	var m model.User
	m.Name = c.FormValue("name")
	m.Email = c.FormValue("email")
	password := c.FormValue("password")

	b, err := model.UserExists(u.DB, m.Email)
	if err != nil {
		return err
	}
	if b {
		c.String(200, "given email address is already used.")
		return nil
	}

	TXHandler(c, u.DB, func(tx *sql.Tx) error {
		if _, err := m.Insert(tx, password); err != nil {
			return err
		}
		return tx.Commit()
	})

	return c.Redirect(301, "/")
}

// Login try login.
func (u *User) Login(c echo.Context) error {
	m, err := model.Auth(u.DB, c.FormValue("email"), c.FormValue("password"))
	if err != nil {
		return err
	}

	log.Printf("authed: %#v", m)

	sess := session.Default(c)
	sess.Set("uid", m.ID)
	sess.Set("name", m.Name)
	sess.Set("email", m.Email)
	sess.Save()

	return c.Redirect(301, "/")
}

// Logout makes user logged out.
func (u *User) Logout(c echo.Context) error {
	sess := session.Default(c)
	sess.Options(session.Options{MaxAge: -1})
	sess.Clear()
	sess.Save()

	// clear cookie
	http.SetCookie(c.Response().Writer, &http.Cookie{
		Name:    "wikisession",
		Value:   "",
		Path:    "/",
		Expires: time.Now().AddDate(0, -1, 0),
	})

	return c.Redirect(301, "/")
}

// LoggedIn returns if current session user is logged in or not.
func LoggedIn(c echo.Context) bool {
	if c == nil {
		return false
	}
	sess := session.Default(c)
	return sess.Get("uid") != nil && sess.Get("name") != nil && sess.Get("email") != nil
}

// CurrentName returns current user name who logged in.
func CurrentName(c echo.Context) string {
	if c == nil {
		return ""
	}
	return session.Default(c).Get("name").(string)
}

// AuthRequired returns a handler function which checks
// if user logged in or not.
func AuthRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !LoggedIn(c) {
			c.Response().WriteHeader(401)
			return nil
		}
		return next(c)
	}
}
