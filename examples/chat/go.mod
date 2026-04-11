module chat

go 1.26

toolchain go1.26.0

require (
	github.com/franchb/htmgo/framework v0.0.0-20260411172522-3bf12b2718d9
	github.com/go-chi/chi/v5 v5.2.5
	github.com/google/uuid v1.6.0
	github.com/mattn/go-sqlite3 v1.14.42
	github.com/puzpuzpuz/xsync/v3 v3.5.1
)

replace github.com/franchb/htmgo/framework => ../../framework
