module github.com/franchb/htmgo/cli/htmgo

go 1.26

require (
	github.com/franchb/htmgo/framework v0.0.0-20260412072145-964b788aa6e0
	github.com/franchb/htmgo/tools/html-to-htmgo v0.0.0-20260412072145-964b788aa6e0
	github.com/fsnotify/fsnotify v1.7.0
	github.com/stretchr/testify v1.11.1
	golang.org/x/mod v0.34.0
	golang.org/x/sys v0.43.0
	golang.org/x/tools v0.43.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
)

require (
	github.com/bmatcuk/doublestar/v4 v4.7.1
	github.com/go-chi/chi/v5 v5.2.5 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/franchb/htmgo/framework => ../../framework

replace github.com/franchb/htmgo/tools/html-to-htmgo => ../../tools/html-to-htmgo
