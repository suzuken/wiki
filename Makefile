DBNAME:=wiki
ENV:=development

deps:
	which godep || go get github.com/tools/godep
	godep restore
	which sql-migrate || go get github.com/rubenv/sql-migrate/...
	which scaneo || go get github.com/variadico/scaneo

run:
	go run ./cmd/wiki/wiki.go

test:
	go test -v ./...

integration-test:
	go test -tags=integration -v ./...

gen:
	cd model && go generate

migrate/init:
	mysql -u root -h localhost --protocol tcp -e "create database \`$(DBNAME)\`" -p

migrate/up:
	sql-migrate up -env=$(ENV)
