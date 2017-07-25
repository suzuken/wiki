package main

import (
	"flag"

	"github.com/suzuken/wiki"
)

func main() {
	var (
		addr   = flag.String("addr", ":8080", "addr to bind")
		dbconf = flag.String("dbconf", "dbconfig.yml", "database configuration file.")
		env    = flag.String("env", "development", "application envirionment (production, development etc.)")
	)
	flag.Parse()
	b := wiki.New()
	b.Init(*dbconf, *env)
	b.Run(*addr)
}
