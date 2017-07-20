package controller

import (
	"database/sql"
	"io"
	"log"
	"net/http"

	"github.com/suzuken/wiki/model"
	"github.com/suzuken/wiki/sessions"
)

// User is controller for requests to user.
type User struct {
	DB *sql.DB
}

// SignUp makes user signup.
func (u *User) SignUp(w http.ResponseWriter, r *http.Request) error {
	var m model.User
	m.Name = r.PostFormValue("name")
	m.Email = r.PostFormValue("email")
	password := r.PostFormValue("password")

	b, err := model.UserExists(u.DB, m.Email)
	if err != nil {
		return err
	}
	if b {
		w.WriteHeader(200)
		io.WriteString(w, "given email address is already used.")
		return nil
	}

	if err := TXHandler(u.DB, func(tx *sql.Tx) error {
		if _, err := m.Insert(tx, password); err != nil {
			return err
		}
		return tx.Commit()
	}); err != nil {
		return err
	}

	http.Redirect(w, r, "/", 301)
	return nil
}

// Login try login.
func (u *User) Login(w http.ResponseWriter, r *http.Request) error {
	m, err := model.Auth(u.DB, r.PostFormValue("email"), r.PostFormValue("password"))
	if err != nil {
		return err
	}

	log.Printf("authed: %#v", m)

	sess, _ := sessions.Get(r, "user")
	sess.Values["id"] = m.ID
	sess.Values["email"] = m.Email
	sess.Values["name"] = m.Name
	if err := sessions.Save(r, w, sess); err != nil {
		log.Printf("session can't save: %s", err)
		return err
	}

	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}

// Logout makes user logged out.
func (u *User) Logout(w http.ResponseWriter, r *http.Request) error {
	sess, _ := sessions.Get(r, "user")
	if err := sessions.Clear(r, w, sess); err != nil {
		return err
	}
	http.Redirect(w, r, "/", 301)
	return nil
}

// LoggedIn returns if current session user is logged in or not.
func LoggedIn(r *http.Request) bool {
	if r == nil {
		return false
	}
	sess, _ := sessions.Get(r, "user")
	id, ok := sess.Values["id"]
	if !ok {
		return false
	}
	return id.(int64) != 0
}

// CurrentName returns current user name who logged in.
func CurrentName(r *http.Request) string {
	if r == nil {
		return ""
	}
	sess, _ := sessions.Get(r, "user")
	rawname, ok := sess.Values["name"]
	if !ok {
		return ""
	}
	name, ok := rawname.(string)
	if !ok {
		return ""
	}
	return name
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
