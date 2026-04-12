package h

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/franchb/htmgo/framework/config"
	"github.com/franchb/htmgo/framework/hx"
	"github.com/franchb/htmgo/framework/service"
)

type RequestContext struct {
	Request           *http.Request
	Response          http.ResponseWriter
	locator           *service.Locator
	isBoosted         bool
	currentBrowserUrl string
	hxPromptResponse  string
	isHxRequest       bool
	hxTargetId        string
	hxTriggerName     string
	hxTriggerId       string
	kv                map[string]interface{}
	parsedQuery       url.Values
}

func GetRequestContext(r *http.Request) *RequestContext {
	val := r.Context().Value(requestContextKey)
	if val == nil {
		return nil
	}
	return val.(*RequestContext)
}

func (c *RequestContext) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response, cookie)
}

func (c *RequestContext) Redirect(path string, code int) {
	if code == 0 {
		code = http.StatusTemporaryRedirect
	}
	if code < 300 || code > 399 {
		code = http.StatusTemporaryRedirect
	}
	c.Response.Header().Set("Location", path)
	c.Response.WriteHeader(code)
}

func (c *RequestContext) IsHttpPost() bool {
	return c.Request.Method == http.MethodPost
}

func (c *RequestContext) IsHttpGet() bool {
	return c.Request.Method == http.MethodGet
}

func (c *RequestContext) IsHttpPut() bool {
	return c.Request.Method == http.MethodPut
}

func (c *RequestContext) IsHttpDelete() bool {
	return c.Request.Method == http.MethodDelete
}

func (c *RequestContext) FormValue(key string) string {
	return c.Request.FormValue(key)
}

func (c *RequestContext) Header(key string) string {
	return c.Request.Header.Get(key)
}

func (c *RequestContext) UrlParam(key string) string {
	return chi.URLParam(c.Request, key)
}

func (c *RequestContext) QueryParam(key string) string {
	if c.parsedQuery == nil {
		c.parsedQuery = c.Request.URL.Query()
	}
	return c.parsedQuery.Get(key)
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
	Router *chi.Mux
}

// Start starts the htmgo server
func Start(opts AppOpts) {
	router := chi.NewRouter()
	instance := App{
		Opts:   opts,
		Router: router,
	}
	instance.start()
}

// requestContextKeyType is an unexported type used as a context key to avoid collisions.
type requestContextKeyType struct{}

var requestContextKey = requestContextKeyType{}

// Deprecated: RequestContextKey is the legacy string context key. Use GetRequestContext instead.
const RequestContextKey = "htmgo.request.context"

func populateHxFields(cc *RequestContext) {
	cc.isBoosted = cc.Request.Header.Get(hx.BoostedHeader) == "true"
	cc.currentBrowserUrl = cc.Request.Header.Get(hx.CurrentUrlHeader)
	cc.hxPromptResponse = cc.Request.Header.Get(hx.PromptResponseHeader)
	cc.isHxRequest = cc.Request.Header.Get(hx.RequestHeader) == "true"
	cc.hxTargetId = cc.Request.Header.Get(hx.TargetIdHeader)
	cc.hxTriggerName = cc.Request.Header.Get(hx.TriggerNameHeader)
	cc.hxTriggerId = cc.Request.Header.Get(hx.TriggerIdHeader)
}

func (app *App) UseWithContext(h func(w http.ResponseWriter, r *http.Request, context map[string]any)) {
	app.Router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cc := GetRequestContext(r)
			if cc == nil {
				handler.ServeHTTP(w, r)
				return
			}
			if cc.kv == nil {
				cc.kv = make(map[string]interface{})
			}
			h(w, r, cc.kv)
			handler.ServeHTTP(w, r)
		})
	})
}

func (app *App) Use(h func(ctx *RequestContext)) {
	app.Router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cc := GetRequestContext(r)
			if cc == nil {
				handler.ServeHTTP(w, r)
				return
			}
			h(cc)
			handler.ServeHTTP(w, r)
		})
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
func StaticCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cacheHeader := "public, max-age=3600"
		if r.URL.RawQuery != "" {
			cacheHeader = "public, max-age=31536000, immutable"
		}
		sw := &cacheControlWriter{ResponseWriter: w, cacheHeader: cacheHeader}
		next.ServeHTTP(sw, r)
	})
}

// cacheControlWriter injects Cache-Control only for successful responses.
type cacheControlWriter struct {
	http.ResponseWriter
	cacheHeader string
}

func (cw *cacheControlWriter) WriteHeader(code int) {
	if code == http.StatusOK || code == http.StatusNotModified {
		cw.ResponseWriter.Header().Set("Cache-Control", cw.cacheHeader)
	}
	cw.ResponseWriter.WriteHeader(code)
}

func (app *App) start() {

	slog.SetLogLoggerLevel(GetLogLevel())

	// Read the configured public asset path so the middleware skip matches
	// what the application actually serves static files under.
	publicPrefix := config.Get().PublicAssetPath
	if publicPrefix == "" {
		publicPrefix = "/public"
	}

	app.Router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip RequestContext creation for static file requests.
			if strings.HasPrefix(r.URL.Path, publicPrefix+"/") || r.URL.Path == publicPrefix {
				h.ServeHTTP(w, r)
				return
			}

			cc := &RequestContext{
				locator:  app.Opts.ServiceLocator,
				Request:  r,
				Response: w,
			}
			populateHxFields(cc)
			ctx := context.WithValue(r.Context(), requestContextKey, cc)
			// Also store with legacy string key for backward compat with generated code
			// and external consumers that use h.RequestContextKey directly.
			ctx = context.WithValue(ctx, RequestContextKey, cc)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
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

	if err := http.ListenAndServe(port, app.Router); err != nil {
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
			if err := http.ListenAndServe(port, app.Router); err != nil {
				slog.Error("Failed to restart server", "error", err)
				panic(err)
			}
		}

		panic(err)
	}
}

func writeHtml(w http.ResponseWriter, element Ren) error {
	if element == nil {
		return nil
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := fmt.Fprint(w, Render(element, WithDocType()))
	return err
}

func HtmlView(w http.ResponseWriter, page *Page) error {
	// if the page is nil, do nothing, this can happen if custom response is written, such as a 302 redirect
	if page == nil {
		return nil
	}
	return writeHtml(w, page.Root)
}

func PartialViewWithHeaders(w http.ResponseWriter, headers *Headers, partial *Partial) error {
	if partial == nil {
		return nil
	}

	if partial.Headers != nil {
		for s, a := range *partial.Headers {
			w.Header().Set(s, a)
		}
	}

	if headers != nil {
		for s, a := range *headers {
			w.Header().Set(s, a)
		}
	}

	return writeHtml(w, partial.Root)
}

func PartialView(w http.ResponseWriter, partial *Partial) error {
	if partial == nil {
		return nil
	}

	if partial.Headers != nil {
		for s, a := range *partial.Headers {
			w.Header().Set(s, a)
		}
	}

	return writeHtml(w, partial.Root)
}
