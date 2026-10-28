package h

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCSPNonce(t *testing.T) {
	t.Parallel()
	a, b := newCSPNonce(), newCSPNonce()
	assert.NotEqual(t, a, b)
	for _, n := range []string{a, b} {
		raw, err := base64.RawURLEncoding.DecodeString(n)
		require.NoError(t, err)
		assert.Len(t, raw, 16)
	}
}

func TestRequestContext_CSPNonceIsStablePerRequest(t *testing.T) {
	t.Parallel()
	rc := &RequestContext{}
	first := rc.CSPNonce()
	assert.NotEmpty(t, first)
	assert.Equal(t, first, rc.CSPNonce())
}

func TestRender_WithScriptNonce(t *testing.T) {
	t.Parallel()
	out := Render(Div(
		Script("/public/app.js"),
		ScriptWithVersion("/public/htmgo.js", "abc"),
		UnsafeRawScript("console.log(1)"),
		Button(OnClick(EvalJsOnParent("console.log('x')"))),
	), WithScriptNonce("n0nce"))

	assert.Contains(t, out, `<script src="/public/app.js" nonce="n0nce">`)
	assert.Contains(t, out, `<script src="/public/htmgo.js?v=abc" nonce="n0nce">`)
	assert.Contains(t, out, `<script nonce="n0nce">console.log(1)</script>`)
	// Covers the helper <script id="__eval_..."> emitted by AddScript.
	assert.Equal(t, strings.Count(out, "<script"), strings.Count(out, `nonce="n0nce"`), out)
}

func TestRender_NoNonceByDefault(t *testing.T) {
	t.Parallel()
	out := Render(Div(Script("/public/app.js"), UnsafeRawScript("x()")))
	assert.NotContains(t, out, "nonce")
	assert.Contains(t, out, `<script>x()</script>`)
}

func TestRender_NonceDoesNotLeakThroughPool(t *testing.T) {
	t.Parallel()
	Render(Div(Script("/a.js")), WithScriptNonce("first"))
	out := Render(Div(Script("/a.js")))
	assert.NotContains(t, out, "first")
	assert.NotContains(t, out, "nonce")
}

func TestUnsafeRawScript_ContentUnchanged(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "<script>x()</script>", UnsafeRawScript("x()").Content)
}

func TestInjectHtmxNonceConfig(t *testing.T) {
	t.Parallel()
	out := injectHtmxNonceConfig(`<!DOCTYPE html><html><head><title>t</title></head></html>`, "abc")
	assert.Equal(t,
		`<!DOCTYPE html><html><head><meta name="htmx-config" content="{&#34;inlineScriptNonce&#34;:&#34;abc&#34;,&#34;inlineStyleNonce&#34;:&#34;abc&#34;}"><title>t</title></head></html>`,
		out)

	existing := `<html><head><meta name="htmx-config" content="{}"></head></html>`
	assert.Equal(t, existing, injectHtmxNonceConfig(existing, "abc"))
}

var nonceRe = regexp.MustCompile(`<script[^>]* nonce="([^"]+)"`)

func newCSPTestApp(policy string) *fiber.App {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(requestContextLocalsKey, &RequestContext{Fiber: c, cspPolicy: policy})
		return c.Next()
	})
	app.Get("/", func(c fiber.Ctx) error {
		return HtmlView(c, NewPage(Html(
			Head(Script("/public/htmgo.js")),
			Body(UnsafeRawScript("x()")),
		)))
	})
	return app
}

func getCSPTestPage(t *testing.T, app *fiber.App) (*http.Response, string) {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, string(body)
}

func TestHtmlView_CSPEnabled(t *testing.T) {
	t.Parallel()
	policy := "script-src 'nonce-{nonce}' 'strict-dynamic'; object-src 'none'"
	app := newCSPTestApp(policy)

	resp, body := getCSPTestPage(t, app)
	m := nonceRe.FindStringSubmatch(body)
	require.NotNil(t, m, body)
	nonce := m[1]

	assert.Equal(t, strings.ReplaceAll(policy, "{nonce}", nonce), resp.Header.Get("Content-Security-Policy"))
	assert.Equal(t, 2, strings.Count(body, `nonce="`+nonce+`"`), body)
	assert.Contains(t, body, `&#34;inlineScriptNonce&#34;:&#34;`+nonce+`&#34;`)

	_, body2 := getCSPTestPage(t, app)
	m2 := nonceRe.FindStringSubmatch(body2)
	require.NotNil(t, m2, body2)
	assert.NotEqual(t, nonce, m2[1], "each request must get a fresh nonce")
}

func TestHtmlView_CSPDisabled(t *testing.T) {
	t.Parallel()
	resp, body := getCSPTestPage(t, newCSPTestApp(""))
	assert.Empty(t, resp.Header.Get("Content-Security-Policy"))
	assert.NotContains(t, body, "nonce")
	assert.NotContains(t, body, "htmx-config")
}

func TestCachedComponent_GetsRequestNonce(t *testing.T) {
	t.Parallel()
	c := Cached(time.Minute, func() *Element { return Div(Script("/public/cached.js")) })

	one := Render(Div(c()), WithScriptNonce("one"))
	two := Render(Div(c()), WithScriptNonce("two"))
	none := Render(Div(c()))

	assert.Contains(t, one, `nonce="one"`)
	assert.Contains(t, two, `nonce="two"`)
	assert.NotContains(t, two, "one")
	assert.NotContains(t, none, "nonce")
	for _, out := range []string{one, two, none} {
		assert.NotContains(t, out, cachedNonceToken)
	}
}

func TestCachedPerKey_GetsRequestNonce(t *testing.T) {
	t.Parallel()
	c := CachedPerKey(time.Minute, func() (string, GetElementFunc) {
		return "k", func() *Element { return Div(Script("/public/k.js")) }
	})

	one := Render(Div(c()), WithScriptNonce("one"))
	two := Render(Div(c()), WithScriptNonce("two"))
	none := Render(Div(c()))

	assert.Contains(t, one, `nonce="one"`)
	assert.Contains(t, two, `nonce="two"`)
	assert.NotContains(t, none, "nonce")
	for _, out := range []string{one, two, none} {
		assert.NotContains(t, out, cachedNonceToken)
	}
}
