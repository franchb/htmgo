module minimal-htmgo

go 1.23.0

require (
	github.com/franchb/htmgo/framework v1.0.7-0.20250703190716-06f01b3d7c1b
	github.com/go-chi/chi/v5 v5.1.0
)

require github.com/google/uuid v1.6.0 // indirect

replace github.com/franchb/htmgo/framework => ../../framework
