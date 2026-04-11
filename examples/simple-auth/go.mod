module simpleauth

go 1.23.0

require (
	github.com/franchb/htmgo/framework v0.0.0-20260411081622-8ee0c30de603
	github.com/mattn/go-sqlite3 v1.14.24
	golang.org/x/crypto v0.28.0
)

require (
	github.com/go-chi/chi/v5 v5.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace github.com/franchb/htmgo/framework => ../../framework
