package h

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/franchb/htmgo/framework/config"
	"github.com/franchb/htmgo/framework/hx"
	"github.com/franchb/htmgo/framework/service"
)

type RequestContext struct {
	Fiber             fiber.Ctx
	locator           *service.Locator
	isBoosted         bool
	currentBrowserUrl string
	hxPromptResponse  string
	isHxRequest       bool
	hxTargetId        string
	hxTriggerName     string
	hxTriggerId       string
	kv                map[string]interface{}
}

// requestContextLocalsKey is the key used to store the RequestContext in Fiber's Locals.
var requestContextLocalsKey = "htmgo.request.context"

func GetRequestContext(c fiber.Ctx) *RequestContext {
	val := c.Locals(requestContextLocalsKey)
	if val == nil {
		return nil
	}
	return val.(*RequestContext)
}

func (c *RequestContext) SetCookie(cookie *fiber.Cookie) {
	c.Fiber.Cookie(cookie)
}

func (c *RequestContext) Redirect(path string, code int) error {
	if code == 0 {
		code = fiber.StatusTemporaryRedirect
	}
	if code < 300 || code > 399 {
		code = fiber.StatusTemporaryRedirect
	}
	return c.Fiber.Redirect().Status(code).To(path)
}

func (c *RequestContext) IsHttpPost() bool {
	return c.Fiber.Method() == fiber.MethodPost
}

func (c *RequestContext) IsHttpGet() bool {
	return c.Fiber.Method() == fiber.MethodGet
}

func (c *RequestContext) IsHttpPut() bool {
	return c.Fiber.Method() == fiber.MethodPut
}

func (c *RequestContext) IsHttpDelete() bool {
	return c.Fiber.Method() == fiber.MethodDelete
}

func (c *RequestContext) FormValue(key string) string {
	return c.Fiber.FormValue(key)
}

func (c *RequestContext) Header(key string) string {
	return c.Fiber.Get(key)
}

func (c *RequestContext) UrlParam(key string) string {
	return c.Fiber.Params(key)
}

func (c *RequestContext) QueryParam(key string) string {
	return c.Fiber.Query(key)
}

func (c *RequestContext) IsBoosted() bool {
	return c.isBoosted
}

func (c *RequestContext) IsHxRequest() bool {
	return c.isHxRequest
}

func (c *RequestContext) HxPromptResponse() string {
	return c.hxPromptResponse
}

func (c *RequestContext) HxTargetId() string {
	return c.hxTargetId
}

func (c *RequestContext) HxTriggerName() string {
	return c.hxTriggerName
}

func (c *RequestContext) HxTriggerId() string {
	return c.hxTriggerId
}

func (c *RequestContext) HxCurrentBrowserUrl() string {
	return c.currentBrowserUrl
}

func (c *RequestContext) Set(key string, value interface{}) {
	if c.kv == nil {
		c.kv = make(map[string]interface{})
	}
	c.kv[key] = value
}

func (c *RequestContext) Get(key string) interface{} {
	if c.kv == nil {
		return nil
	}
	return c.kv[key]
}

// ServiceLocator returns the service locator to register and retrieve services
// Usage:
// service.Set[db.Queries](locator, service.Singleton, db.Provide)
// service.Get[db.Queries](locator)
func (c *RequestContext) ServiceLocator() *service.Locator {
	return c.locator
}

type AppOpts struct {
	LiveReload     bool
	ServiceLocator *service.Locator
	Register       func(app *App)
	Port           int
}

type App struct {
	Opts   AppOpts
	Router *fiber.App
}

// Start starts the htmgo server
func Start(opts AppOpts) {
	fiberApp := fiber.New()
	instance := App{
		Opts:   opts,
		Router: fiberApp,
	}
	instance.start()
}

func populateHxFields(cc *RequestContext) {
	cc.isBoosted = cc.Fiber.Get(hx.BoostedHeader) == "true"
	cc.currentBrowserUrl = cc.Fiber.Get(hx.CurrentUrlHeader)
	cc.hxPromptResponse = cc.Fiber.Get(hx.PromptResponseHeader)
	cc.isHxRequest = cc.Fiber.Get(hx.RequestHeader) == "true"
	cc.hxTargetId = cc.Fiber.Get(hx.TargetIdHeader)
	cc.hxTriggerName = cc.Fiber.Get(hx.TriggerNameHeader)
	cc.hxTriggerId = cc.Fiber.Get(hx.TriggerIdHeader)
}

func (app *App) Use(h func(ctx *RequestContext)) {
	app.Router.Use(func(c fiber.Ctx) error {
		cc := GetRequestContext(c)
		if cc == nil {
			return c.Next()
		}
		h(cc)
		return c.Next()
	})
}

func GetLogLevel() slog.Level {
	// Get the log level from the environment variable
	logLevel := os.Getenv("LOG_LEVEL")
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		// Default to INFO if no valid log level is set
		return slog.LevelInfo
	}
}

// StaticCacheMiddleware adds Cache-Control headers for static file requests.
// Requests with a query string (e.g. ?v=hash) are treated as immutable with a
// one-year max-age; all other static requests get a one-hour max-age.
// Headers are only applied to successful responses (2xx/304) so that transient
// 404s or errors are not cached by browsers or CDNs.
func StaticCacheMiddleware(c fiber.Ctx) error {
	cacheHeader := "public, max-age=3600"
	if len(c.Request().URI().QueryString()) > 0 {
		cacheHeader = "public, max-age=31536000, immutable"
	}

	// Call next handler first, then set cache header based on status code.
	err := c.Next()

	status := c.Response().StatusCode()
	if status == fiber.StatusOK || status == fiber.StatusNotModified {
		c.Set("Cache-Control", cacheHeader)
	}

	return err
}

func (app *App) start() {

	slog.SetLogLoggerLevel(GetLogLevel())

	// Read the configured public asset path so the middleware skip matches
	// what the application actually serves static files under.
	publicPrefix := config.Get().PublicAssetPath
	if publicPrefix == "" {
		publicPrefix = "/public"
	}

	app.Router.Use(func(c fiber.Ctx) error {
		// Skip RequestContext creation for static file requests.
		path := c.Path()
		if strings.HasPrefix(path, publicPrefix+"/") || path == publicPrefix {
			return c.Next()
		}

		cc := &RequestContext{
			locator: app.Opts.ServiceLocator,
			Fiber:   c,
		}
		populateHxFields(cc)
		c.Locals(requestContextLocalsKey, cc)
		return c.Next()
	})

	if app.Opts.Register != nil {
		app.Opts.Register(app)
	}

	if app.Opts.LiveReload && IsDevelopment() {
		app.AddLiveReloadHandler("/dev/livereload")
	}

	port := ":3000"
	isDefaultPort := true

	if os.Getenv("PORT") != "" {
		port = fmt.Sprintf(":%s", os.Getenv("PORT"))
		isDefaultPort = false
	}

	if app.Opts.Port != 0 {
		port = fmt.Sprintf(":%d", app.Opts.Port)
		isDefaultPort = false
	}

	if isDefaultPort {
		slog.Info("Using default port 3000, set PORT environment variable to change it or use AppOpts.Port")
	}

	slog.Info(fmt.Sprintf("Server started at localhost%s", port))

	if err := app.Router.Listen(port); err != nil {
		// If we are in watch mode, just try to kill any processes holding that port
		// and try again
		if IsDevelopment() && IsWatchMode() {
			slog.Info("Port already in use, trying to kill the process and start again")
			if runtime.GOOS == "windows" {
				cmd := exec.Command("cmd", "/C", fmt.Sprintf(`for /F "tokens=5" %%i in ('netstat -aon ^| findstr :%s') do taskkill /F /PID %%i`, port))
				cmd.Run()
			} else {
				cmd := exec.Command("bash", "-c", fmt.Sprintf("kill -9 $(lsof -ti%s)", port))
				cmd.Run()
			}

			time.Sleep(time.Millisecond * 50)

			// Try to start server again
			if err := app.Router.Listen(port); err != nil {
				slog.Error("Failed to restart server", "error", err)
				panic(err)
			}
		}

		panic(err)
	}
}

func writeHtml(c fiber.Ctx, element Ren) error {
	if element == nil {
		return nil
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(Render(element, WithDocType()))
}

func HtmlView(c fiber.Ctx, page *Page) error {
	// if the page is nil, do nothing, this can happen if custom response is written, such as a 302 redirect
	if page == nil {
		return nil
	}
	return writeHtml(c, page.Root)
}

func PartialViewWithHeaders(c fiber.Ctx, headers *Headers, partial *Partial) error {
	if partial == nil {
		return nil
	}

	if partial.Headers != nil {
		for s, a := range *partial.Headers {
			c.Set(s, a)
		}
	}

	if headers != nil {
		for s, a := range *headers {
			c.Set(s, a)
		}
	}

	return writeHtml(c, partial.Root)
}

func PartialView(c fiber.Ctx, partial *Partial) error {
	if partial == nil {
		return nil
	}

	if partial.Headers != nil {
		for s, a := range *partial.Headers {
			c.Set(s, a)
		}
	}

	return writeHtml(c, partial.Root)
}
