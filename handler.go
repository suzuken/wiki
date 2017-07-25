package wiki

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/suzuken/wiki/controller"
	"github.com/suzuken/wiki/httputil"
)

var errUnauthrized = errors.New("unauthorized")

// Auth verify if session user is logged in.
func Auth(h handler) handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if !controller.LoggedIn(r) {
			return &httputil.HTTPError{
				Status: http.StatusUnauthorized,
				Err:    errUnauthrized,
			}
		}
		h.ServeHTTP(w, r)
		return nil
	}
}

func m(method string, h handler) handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != method {
			return &httputil.HTTPError{Status: http.StatusMethodNotAllowed}
		}
		h.ServeHTTP(w, r)
		return nil
	}
}

func GET(h handler) handler  { return m("GET", h) }
func POST(h handler) handler { return m("POST", h) }

type handler func(w http.ResponseWriter, r *http.Request) error

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	runHandler(w, r, h, handleError)
}

type errFn func(w http.ResponseWriter, r *http.Request, status int, err error)

func logError(req *http.Request, err error, rv interface{}) {
	if err != nil {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "Error serving %s: %v\n", req.URL, err)
		if rv != nil {
			fmt.Fprintln(&buf, rv)
			buf.Write(debug.Stack())
		}
		log.Print(buf.String())
	}
}

func runHandler(w http.ResponseWriter, r *http.Request,
	fn func(w http.ResponseWriter, r *http.Request) error, errfn errFn) {
	defer func() {
		if rv := recover(); rv != nil {
			err := errors.New("handler panic")
			logError(r, err, rv)
			errfn(w, r, http.StatusInternalServerError, err)
		}
	}()

	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	r.ParseForm()
	var buf httputil.ResponseBuffer
	err := fn(&buf, r)
	if err == nil {
		buf.WriteTo(w)
	} else if e, ok := err.(*httputil.HTTPError); ok {
		if e.Status >= 500 {
			logError(r, err, nil)
		}
		errfn(w, r, e.Status, e.Err)
	} else {
		logError(r, err, nil)
		errfn(w, r, http.StatusInternalServerError, err)
	}
}

func errorText(err error) string {
	if err == errUnauthrized {
		return "You are unauthorized."
	}
	return "Internal Server error."
}

func handleError(w http.ResponseWriter, r *http.Request, status int, err error) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	io.WriteString(w, errorText(err))
}
