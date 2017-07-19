package sessions

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// store is session store based on gorilla/sessions.
// This is singleton for wiki app.
var store = sessions.NewCookieStore([]byte("secretkey"))

func Get(r *http.Request, key string) (*sessions.Session, error) {
	return store.Get(r, key)
}

func Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return store.Save(r, w, session)
}

// Clear removes the given session
func Clear(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	session.Options.MaxAge = -1
	return Save(r, w, session)
}
