module htmgo-site

go 1.27.0

require (
	github.com/alecthomas/chroma/v2 v2.27.0
	github.com/franchb/htmgo/framework/v2 v2.1.5-0.20261123191958-07270a1d05fe
	github.com/franchb/htmgo/tools/html-to-htmgo/v2 v2.1.4-0.20261123191958-07270a1d05fe
	github.com/gofiber/fiber/v3 v3.7.0
	github.com/google/uuid v1.6.0
	github.com/yuin/goldmark v1.8.6
	github.com/yuin/goldmark-highlighting/v2 v2.0.0-20230729083705-37449abec8cc
)

require (
	github.com/dlclark/regexp2/v2 v2.2.1 // indirect
	github.com/gofiber/schema v1.9.3 // indirect
	github.com/gofiber/utils/v2 v2.6.4 // indirect
	github.com/klauspost/compress v1.20.4 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/molecule-man/go-brrr v1.3.1 // indirect
	github.com/philhofer/fwd v1.2.1 // indirect
	github.com/tinylib/msgp v1.6.5 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.79.0 // indirect
	golang.org/x/crypto v0.59.0 // indirect
	golang.org/x/net v0.63.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/text v0.44.0 // indirect
	golang.org/x/tools v0.53.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/franchb/htmgo/framework/v2 => ../framework

replace github.com/franchb/htmgo/tools/html-to-htmgo/v2 => ../tools/html-to-htmgo
