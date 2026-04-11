module minimal-htmgo

go 1.23.0

require (
	github.com/franchb/htmgo/framework v0.0.0-20260411081622-8ee0c30de603
	github.com/go-chi/chi/v5 v5.1.0
)

require github.com/google/uuid v1.6.0 // indirect

replace github.com/franchb/htmgo/framework => ../../framework
