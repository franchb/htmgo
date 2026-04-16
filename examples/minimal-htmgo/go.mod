module minimal-htmgo

go 1.26

toolchain go1.26.0

require (
	github.com/franchb/htmgo/framework v0.0.0-20260416123109-28119a474a28
	github.com/go-chi/chi/v5 v5.2.5
)

require (
	github.com/google/uuid v1.6.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/franchb/htmgo/framework => ../../framework
