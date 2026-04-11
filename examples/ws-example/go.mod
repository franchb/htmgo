module ws-example

go 1.23.0

require (
	github.com/franchb/htmgo/extensions/websocket v0.0.0-00010101000000-000000000000
	github.com/franchb/htmgo/framework v0.0.0-20260411081622-8ee0c30de603
)

require (
	github.com/go-chi/chi/v5 v5.1.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/puzpuzpuz/xsync/v3 v3.4.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
)

replace github.com/franchb/htmgo/framework => ../../framework

replace github.com/franchb/htmgo/extensions/websocket => ../../extensions/websocket
