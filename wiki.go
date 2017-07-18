package main

import (
	"database/sql"
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ipfans/echo-session"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/suzuken/wiki/controller"
	"github.com/suzuken/wiki/db"
)

// Server is whole server implementation for this wiki app.
// This holds database connection and router settings based on labstack/echo.
type Server struct {
	db     *sql.DB
	Engine *echo.Echo
}

// Close makes the database connection to close.
func (s *Server) Close() error {
	return s.db.Close()
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
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
	s.db = db

	// NOTE: define helper func to use from templates here.
	t := &Template{
		templates: template.Must(template.New("").Funcs(template.FuncMap{
			"LoggedIn":    controller.LoggedIn,
			"CurrentName": controller.CurrentName,
		}).ParseGlob("templates/*")),
	}
	s.Engine.Renderer = t

	store := session.NewCookieStore([]byte("secretkey"))
	s.Engine.Use(session.Sessions("wikisession", store))
	s.Engine.Use(middleware.CSRF())
	s.Route()
}

func CSRFToken(c echo.Context) string {
	return c.Get(middleware.DefaultCSRFConfig.ContextKey).(string)
}

// New returns server object.
func New() *Server {
	return &Server{Engine: echo.New()}
}

// Run starts running http server.
func (s *Server) Run(addr string) {
	s.Engine.Start(addr)
}

// Route setting router for this wiki.
func (s *Server) Route() {
	article := &controller.Article{DB: s.db}
	user := &controller.User{DB: s.db}

	auth := s.Engine.Group("/")
	auth.Use(controller.AuthRequired)
	{
		auth.GET("/authtest", func(c echo.Context) error {
			return c.String(200, "you're authed")
		})
		auth.GET("/new", func(c echo.Context) error {
			return c.Render(200, "new.tmpl", echo.Map{
				"title":   "New: go-wiki",
				"csrf":    CSRFToken(c),
				"context": c,
			})
		})
		auth.GET("/article/:id/edit", article.Edit)
		auth.POST("/save", article.Save)
		auth.POST("/delete", article.Delete)
		auth.GET("/logout", func(c echo.Context) error {
			return c.Render(http.StatusOK, "logout.tmpl", echo.Map{
				"csrf":    CSRFToken(c),
				"context": c,
			})
		})
		auth.POST("/logout", user.Logout)
	}

	s.Engine.GET("/", article.Root)
	s.Engine.GET("/article/:id", article.Get)
	s.Engine.GET("/signup", func(c echo.Context) error {
		return c.Render(http.StatusOK, "signup.tmpl", echo.Map{
			"csrf": CSRFToken(c),
		})
	})
	s.Engine.POST("/signup", user.SignUp)
	s.Engine.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.tmpl", echo.Map{
			"csrf": CSRFToken(c),
		})
	})
	s.Engine.POST("/login", user.Login)

	s.Engine.Static("/static", "static")
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
