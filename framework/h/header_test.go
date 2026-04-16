package h

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/franchb/htmgo/framework/hx"
)

func TestReplaceUrlHeader(t *testing.T) {
	headers := ReplaceUrlHeader("/new-url")
	assert.Contains(t, *headers, hx.ReplaceUrlHeader)
	assert.Equal(t, "/new-url", (*headers)[hx.ReplaceUrlHeader])
}

func TestPushUrlHeader(t *testing.T) {
	headers := PushUrlHeader("/push-url")
	assert.Contains(t, *headers, hx.PushUrlHeader)
	assert.Equal(t, "/push-url", (*headers)[hx.PushUrlHeader])
}

func TestPushQsHeader(t *testing.T) {
	ctx := &RequestContext{currentBrowserUrl: "https://example.com/path"}
	qs := NewQs("a", "b", "c", "d")
	headers := PushQsHeader(ctx, qs)
	expectedURL := "/path?a=b&c=d"
	assert.Contains(t, *headers, hx.ReplaceUrlHeader)
	assert.Equal(t, expectedURL, (*headers)[hx.ReplaceUrlHeader])
}

func TestCombineHeaders(t *testing.T) {
	h1 := NewHeaders("Content-Type", "application/json")
	h2 := NewHeaders("Authorization", "Bearer token")
	combined := CombineHeaders(h1, h2)
	assert.Equal(t, "application/json", (*combined)["Content-Type"])
	assert.Equal(t, "Bearer token", (*combined)["Authorization"])
}

func TestCurrentPath(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	fctx.Request.Header.Set(hx.CurrentUrlHeader, "https://example.com/current-path")
	c := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(c)

	ctx := &RequestContext{Fiber: c}
	path := CurrentPath(ctx)
	assert.Equal(t, "/current-path", path)
}

func TestNewHeaders(t *testing.T) {
	headers := NewHeaders("X-Custom", "value", "X-Another", "another-value")
	require.NotNil(t, headers)
	assert.Equal(t, "value", (*headers)["X-Custom"])
	assert.Equal(t, "another-value", (*headers)["X-Another"])

	invalidHeaders := NewHeaders("X-Custom")
	assert.Empty(t, *invalidHeaders) // Check incorrect pair length handling
}
