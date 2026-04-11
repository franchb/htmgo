module htmgo-site

go 1.23.0

require (
	github.com/alecthomas/chroma/v2 v2.14.0
	github.com/franchb/htmgo/framework v0.0.0-20260411081622-8ee0c30de603
	github.com/franchb/htmgo/tools/html-to-htmgo v0.0.0-20260411081622-8ee0c30de603
	github.com/go-chi/chi/v5 v5.1.0
	github.com/google/uuid v1.6.0
	github.com/yuin/goldmark v1.7.4
	github.com/yuin/goldmark-highlighting/v2 v2.0.0-20230729083705-37449abec8cc
)

require (
	github.com/dlclark/regexp2 v1.11.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
)

replace github.com/franchb/htmgo/framework => ../framework

replace github.com/franchb/htmgo/tools/html-to-htmgo => ../tools/html-to-htmgo
