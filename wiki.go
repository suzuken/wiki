package wiki

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/suzuken/wiki/controller"
	"github.com/suzuken/wiki/db"
	"github.com/suzuken/wiki/httputil"
	"github.com/suzuken/wiki/view"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/context"
)

// Server is whole server implementation for this wiki app.
// This holds database connection and router settings based on gin.
type Server struct {
	db  *sql.DB
	mux *http.ServeMux
}

// Close makes the database connection to close.
func (s *Server) Close() error {
	return s.db.Close()
}

// Init initialize server state. Connecting to database, compiling templates,
// and settings router.
func (s *Server) Init(dbconf, env string) {
	cs, err := db.NewConfigsFromFile(dbconf)
	if err != nil {
		log.Fatalf("cannot open database configuration. exit. %s", err)
	}
	db, err := cs.Open(env)
	if err != nil {
		log.Fatalf("db initialization failed: %s", err)
	}

	view.Funcs(template.FuncMap{
		"LoggedIn":    controller.LoggedIn,
		"CurrentName": controller.CurrentName,
	})
	view.Init()

	s.db = db
	s.Route()
}

// New returns server object.
func New() *Server {
	return &Server{}
}

// Run starts running http server.
func (s *Server) Run(addr string) {
	log.Printf("start listening on %s", addr)
	http.ListenAndServe(addr, context.ClearHandler(s.mux))
}

func Auth(h handler) handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if !controller.LoggedIn(r) {
			return &httputil.HTTPError{Status: http.StatusUnauthorized}
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
	return "Internal Server error."
}

func handleError(w http.ResponseWriter, r *http.Request, status int, err error) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	io.WriteString(w, errorText(err))
}

// Route setting router for this wiki.
func (s *Server) Route() {
	mux := http.NewServeMux()

	article := &controller.Article{DB: s.db}
	user := &controller.User{DB: s.db}

	mux.Handle("/authtest", GET(Auth(controller.AuthTestHandler)))
	mux.Handle("/new", GET(controller.NewArticleHandler))
	mux.Handle("/article/", GET(article.Get))
	mux.Handle("/article/edit/", GET(Auth(article.Edit)))
	mux.Handle("/save", POST(Auth(article.Save)))
	mux.Handle("/delete", POST(Auth(article.Delete)))
	mux.Handle("/logout", handler(user.LogoutHandler))

	mux.Handle("/", GET(article.Root))
	mux.Handle("/signup", handler(user.SignupHandler))
	mux.Handle("/login", handler(user.LoginHandler))
	mux.Handle("/static", http.FileServer(http.Dir("./static")))
	s.mux = mux
}
