package controller

import (
	"database/sql"
	"io"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/suzuken/wiki/model"
	"github.com/suzuken/wiki/sessions"
)

// User is controller for requests to user.
type User struct {
	DB *sql.DB
}

// SignUp makes user signup.
func (u *User) SignUp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var m model.User
	m.Name = r.PostFormValue("name")
	m.Email = r.PostFormValue("email")
	password := r.PostFormValue("password")

	b, err := model.UserExists(u.DB, m.Email)
	if err != nil {
		log.Printf("query error: %s", err)
		http.Error(w, "db error", 500)
		return
	}
	if b {
		w.WriteHeader(200)
		io.WriteString(w, "given email address is already used.")
		return
	}

	if err := TXHandler(u.DB, func(tx *sql.Tx) error {
		if _, err := m.Insert(tx, password); err != nil {
			return err
		}
		return tx.Commit()
	}); err != nil {
		Error(w, err, 500)
		return
	}

	http.Redirect(w, r, "/", 301)
}

// Login try login.
func (u *User) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	m, err := model.Auth(u.DB, r.PostFormValue("email"), r.PostFormValue("password"))
	if err != nil {
		log.Printf("auth failed: %s", err)
		Error(w, err, 500)
		return
	}

	log.Printf("authed: %#v", m)

	sess, _ := sessions.Get(r, "user")
	sess.Values["user"] = m.Mask()
	sessions.Save(r, w, sess)

	http.Redirect(w, r, "/", 301)
}

// Logout makes user logged out.
func (u *User) Logout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sess, _ := sessions.Get(r, "user")
	sessions.Clear(r, w, sess)
	http.Redirect(w, r, "/", 301)
}

// LoggedIn returns if current session user is logged in or not.
func LoggedIn(r *http.Request) bool {
	if r == nil {
		return false
	}
	sess, _ := sessions.Get(r, "user")
	rawuser, ok := sess.Values["user"]
	if !ok {
		return false
	}
	user, ok := rawuser.(model.User)
	if !ok {
		return false
	}
	return user.ID != 0 && user.Name != "" && user.Email != ""
}

// CurrentName returns current user name who logged in.
func CurrentName(r *http.Request) string {
	if r == nil {
		return ""
	}
	sess, _ := sessions.Get(r, "user")
	rawuser, ok := sess.Values["user"]
	if !ok {
		return ""
	}
	user, ok := rawuser.(model.User)
	if !ok {
		return ""
	}
	return user.Name
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
