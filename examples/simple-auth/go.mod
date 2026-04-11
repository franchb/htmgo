module simpleauth

go 1.26

toolchain go1.26.0

require (
	github.com/franchb/htmgo/framework v0.0.0-20260411081622-8ee0c30de603
	github.com/mattn/go-sqlite3 v1.14.42
	golang.org/x/crypto v0.50.0
)

require (
	github.com/go-chi/chi/v5 v5.2.5 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace github.com/franchb/htmgo/framework => ../../framework
