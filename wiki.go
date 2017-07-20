package main

import (
	"database/sql"
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/suzuken/wiki/controller"
	"github.com/suzuken/wiki/db"
	"github.com/suzuken/wiki/view"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
	"github.com/julienschmidt/httprouter"
)

// Server is whole server implementation for this wiki app.
// This holds database connection and router settings based on gin.
type Server struct {
	db  *sql.DB
	mux *httprouter.Router
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
	s.mux = httprouter.New()
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

func Auth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !controller.LoggedIn(r) {
			http.Error(w, "abort", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	}
}

func AuthParam(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if !controller.LoggedIn(r) {
			http.Error(w, "abort", http.StatusUnauthorized)
			return
		}
		h(w, r, p)
	}
}

// Route setting router for this wiki.
func (s *Server) Route() {
	article := &controller.Article{DB: s.db}
	user := &controller.User{DB: s.db}

	s.mux.Handler("GET", "/authtest", Auth(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		io.WriteString(w, "your're authed")
	}))
	s.mux.HandlerFunc("GET", "/new", func(w http.ResponseWriter, r *http.Request) {
		view.HTML(w, 200, "new.tmpl", map[string]interface{}{
			"title":          "New: go-wiki",
			csrf.TemplateTag: csrf.TemplateField(r),
			"request":        r,
		})
	})
	s.mux.GET("/article/:id/edit", AuthParam(article.Edit))
	s.mux.POST("/save", AuthParam(article.Save))
	s.mux.POST("/delete", AuthParam(article.Delete))
	s.mux.HandlerFunc("GET", "/logout", func(w http.ResponseWriter, r *http.Request) {
		view.HTML(w, http.StatusOK, "logout.tmpl", map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(r),
			"request":        r,
		})
	})
	s.mux.HandlerFunc("POST", "/logout", user.Logout)

	s.mux.GET("/", article.Root)
	s.mux.GET("/article/:id", article.Get)
	s.mux.HandlerFunc("GET", "/signup", func(w http.ResponseWriter, r *http.Request) {
		view.HTML(w, http.StatusOK, "signup.tmpl", map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(r),
		})
	})
	s.mux.HandlerFunc("POST", "/signup", user.SignUp)
	s.mux.HandlerFunc("GET", "/login", func(w http.ResponseWriter, r *http.Request) {
		view.HTML(w, http.StatusOK, "login.tmpl", map[string]interface{}{
			csrf.TemplateTag: csrf.TemplateField(r),
		})
	})
	s.mux.HandlerFunc("POST", "/login", user.Login)

	s.mux.ServeFiles("/static/*filepath", http.Dir("./static"))
}

func main() {
	var (
		addr   = flag.String("addr", ":8080", "addr to bind")
		dbconf = flag.String("dbconf", "dbconfig.yml", "database configuration file.")
		env    = flag.String("env", "development", "application envirionment (production, development etc.)")
	)
	flag.Parse()
	b := New()
	b.Init(*dbconf, *env)
	b.Run(*addr)
}
