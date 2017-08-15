# wiki

Example implementation of wiki.

## How to run

    go get github.com/suzuken/wiki/cmd/wiki
    wiki

or

    make run

## Requirements

* Go 1.7 or later
* MySQL 5.6

## Tips

### Generate Scans

Using [scaneo](https://github.com/variadico/scaneo). Simply,

    # edit models, and
    make gen

### DB

Use docker container. For database migration, [sql-migrate](https://github.com/rubenv/sql-migrate)

    # Docker's MySQL build
    make docker/build
    make docker/start

    # run migrate/up after adding ddl in migrations dir.
    make migrate/up

Originally from [gin-boilerplate](https://github.com/voyagegroup/gin-boilerplate)

## Author

Kenta SUZUKI a.k.a. suzuken

## LICENSE

MIT
