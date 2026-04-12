module github.com/franchb/htmgo/extensions/websocket

go 1.26

toolchain go1.26.0

require (
	github.com/franchb/htmgo/framework v0.0.0-20260412023854-358a61b926ff
	github.com/gobwas/ws v1.4.0
	github.com/puzpuzpuz/xsync/v3 v3.5.1
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-chi/chi/v5 v5.2.5 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/franchb/htmgo/framework => ../../framework
