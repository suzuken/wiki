package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"

	"github.com/suzuken/wiki/controller"
	"github.com/suzuken/wiki/db"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// Server is whole server implementation for this wiki app.
// This holds database connection and router settings based on gin.
type Server struct {
	db     *sql.DB
	Engine *gin.Engine
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
	s.db = db
	s.Engine.LoadHTMLGlob("templates/*")
	s.Route()
}

// New returns server object.
func New() *Server {
	r := gin.Default()
	return &Server{Engine: r}
}

// Run starts running http server.
func (s *Server) Run(addr ...string) {
	s.Engine.Run(addr...)
}

// Route setting router for this wiki.
func (s *Server) Route() {
	article := &controller.Article{DB: s.db}

	s.Engine.GET("/", article.Root)
	s.Engine.GET("/article/:id", article.Get)
	s.Engine.GET("/new", func(c *gin.Context) {
		c.HTML(http.StatusOK, "new.tmpl", gin.H{
			"title": "New: go-wiki",
		})
	})
	s.Engine.GET("/article/:id/edit", article.Edit)
	s.Engine.POST("/save", article.Save)
	s.Engine.POST("/delete", article.Delete)
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
