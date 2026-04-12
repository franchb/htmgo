module htmgo-site

go 1.26

toolchain go1.26.0

require (
	github.com/alecthomas/chroma/v2 v2.23.1
	github.com/franchb/htmgo/framework v0.0.0-20260412023854-358a61b926ff
	github.com/franchb/htmgo/tools/html-to-htmgo v0.0.0-20260412023854-358a61b926ff
	github.com/go-chi/chi/v5 v5.2.5
	github.com/google/uuid v1.6.0
	github.com/yuin/goldmark v1.8.2
	github.com/yuin/goldmark-highlighting/v2 v2.0.0-20230729083705-37449abec8cc
)

require (
	github.com/dlclark/regexp2 v1.11.5 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/tools v0.43.0 // indirect
)

replace github.com/franchb/htmgo/framework => ../framework

replace github.com/franchb/htmgo/tools/html-to-htmgo => ../tools/html-to-htmgo
