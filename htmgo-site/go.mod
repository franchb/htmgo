module htmgo-site

go 1.26

toolchain go1.26.0

require (
	github.com/alecthomas/chroma/v2 v2.23.1
	github.com/franchb/htmgo/framework/v2 v2.0.5-0.20260805125541-0e47199aea6e
	github.com/franchb/htmgo/tools/html-to-htmgo/v2 v2.1.1-0.20260805125541-0e47199aea6e
	github.com/gofiber/fiber/v3 v3.4.0
	github.com/google/uuid v1.6.0
	github.com/yuin/goldmark v1.8.2
	github.com/yuin/goldmark-highlighting/v2 v2.0.0-20230729083705-37449abec8cc
)

require (
	github.com/andybalholm/brotli v1.2.2 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/gofiber/schema v1.8.0 // indirect
	github.com/gofiber/utils/v2 v2.1.1 // indirect
	github.com/klauspost/compress v1.19.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/tinylib/msgp v1.6.4 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.72.0 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	golang.org/x/tools v0.47.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/franchb/htmgo/framework/v2 => ../framework

replace github.com/franchb/htmgo/tools/html-to-htmgo/v2 => ../tools/html-to-htmgo
