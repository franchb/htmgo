package h

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

// TestInjectLivereloadMeta_DevMode verifies the meta marker is present in
// the rendered page when IsDevelopment() returns true, and absent otherwise.
func TestInjectLivereloadMeta_DevMode(t *testing.T) {
	page := NewPage(
		Html(
			Head(
				Meta("viewport", "width=device-width, initial-scale=1"),
			),
			Body(Text("hello")),
		),
	)

	t.Run("dev mode injects meta", func(t *testing.T) {
		t.Setenv("ENV", "development")
		resetEnvCache()
		defer func() {
			t.Setenv("ENV", "")
			resetEnvCache()
		}()

		rendered := Render(page.Root, WithDocType())
		injected := injectLivereloadMeta(rendered)
		assert.Contains(t, injected, `<meta name="htmgo-livereload"`)
		assert.True(t, strings.Index(injected, `<head>`)+len(`<head>`) ==
			strings.Index(injected, `<meta name="htmgo-livereload"`),
			"meta must be the first child of <head>")
	})

	t.Run("non-dev mode does not inject meta", func(t *testing.T) {
		t.Setenv("ENV", "")
		resetEnvCache()

		rendered := Render(page.Root, WithDocType())
		assert.NotContains(t, rendered, "htmgo-livereload")
		// injectLivereloadMeta itself always injects; the guard is IsDevelopment() in HtmlView.
		// This sub-test confirms the raw render has no marker.
	})
}

// TestHtmlView_InjectsLivereloadMetaInDev verifies that HtmlView produces the
// meta tag in the response body when ENV=development.
func TestHtmlView_InjectsLivereloadMetaInDev(t *testing.T) {
	t.Setenv("ENV", "development")
	resetEnvCache()
	defer func() {
		t.Setenv("ENV", "")
		resetEnvCache()
	}()

	fiberApp := fiber.New()
	fiberApp.Get("/", func(c fiber.Ctx) error {
		page := NewPage(Html(Head(), Body(Text("world"))))
		return HtmlView(c, page)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := fiberApp.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	buf := make([]byte, 4096)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])
	assert.Contains(t, body, `<meta name="htmgo-livereload"`)
}

// TestHtmlView_NoMetaInProd verifies that HtmlView does NOT inject the
// livereload meta tag when ENV is not "development".
func TestHtmlView_NoMetaInProd(t *testing.T) {
	t.Setenv("ENV", "production")
	resetEnvCache()
	defer func() {
		t.Setenv("ENV", "")
		resetEnvCache()
	}()

	fiberApp := fiber.New()
	fiberApp.Get("/", func(c fiber.Ctx) error {
		page := NewPage(Html(Head(), Body(Text("world"))))
		return HtmlView(c, page)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := fiberApp.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	buf := make([]byte, 4096)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])
	assert.NotContains(t, body, "htmgo-livereload")
}

func TestRequestContext_HxSource(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	called := false
	app.Get("/", func(c fiber.Ctx) error {
		called = true
		rc := &RequestContext{Fiber: c}
		populateHxFields(rc)
		assert.Equal(t, "button#save", rc.HxSource())
		assert.Equal(t, "save", rc.HxSourceID())
		assert.Equal(t, "partial", rc.HxRequestType())
		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HX-Source", "button#save")
	req.Header.Set("HX-Request-Type", "partial")
	_, err := app.Test(req)
	assert.NoError(t, err)
	assert.True(t, called, "handler must be invoked")
}
