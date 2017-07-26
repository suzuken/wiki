package wiki

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/suzuken/wiki/controller"
	"github.com/suzuken/wiki/db"
	"github.com/suzuken/wiki/view"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
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
func (s *Server) Init(dbconf, env string, debug bool) {
	cs, err := db.NewConfigsFromFile(dbconf)
	if err != nil {
		log.Fatalf("cannot open database configuration. exit. %s", err)
	}
	db, err := cs.Open(env)
	if err != nil {
		log.Fatalf("db initialization failed: %s", err)
	}

	// In debug mode, we compile templates on every request.
	view.Init(template.FuncMap{
		"LoggedIn":    controller.LoggedIn,
		"CurrentName": controller.CurrentName,
	}, debug)

	s.db = db
	s.Route()
}

// New returns server object.
func New() *Server {
	return &Server{}
}

// csrfProtectKey should have 32 byte length.
var csrfProtectKey = []byte("32-byte-long-auth-key")

// Run starts running http server.
func (s *Server) Run(addr string) {
	log.Printf("start listening on %s", addr)

	// NOTE: when you serve on TLS, make csrf.Secure(true)
	CSRF := csrf.Protect(
		csrfProtectKey, csrf.Secure(false))
	http.ListenAndServe(addr, context.ClearHandler(CSRF(s.mux)))
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
