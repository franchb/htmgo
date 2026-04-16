package h

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func assertHas(t *testing.T, qs *Qs, key string, value string) {
	str := qs.ToString()
	if value == "" {
		assert.Contains(t, str, key)
		assert.NotContains(t, str, key+"=")
	} else {
		assert.Contains(t, str, key+"="+value)
	}
}

func TestQs(t *testing.T) {
	t.Parallel()
	qs := NewQs("a", "b", "c")
	assertHas(t, qs, "a", "b")
	assertHas(t, qs, "c", "")

	qs2 := NewQs("a", "b", "c", "d")
	assertHas(t, qs2, "a", "b")
	assertHas(t, qs2, "c", "d")

	qs2.Add("e", "f")
	assertHas(t, qs2, "a", "b")
	assertHas(t, qs2, "c", "d")
	assertHas(t, qs2, "e", "f")

	qs2.Remove("e")
	assert.NotContains(t, qs2.ToString(), "e")
}

func TestSetQsOnUrl(t *testing.T) {
	t.Parallel()
	qs := NewQs("a", "b", "c", "d")
	set := SetQueryParams("https://example.com/path", qs)
	assert.Equal(t, "https://example.com/path?a=b&c=d", set)
}

func TestSetQsOnUrlWithDelete(t *testing.T) {
	t.Parallel()
	qs := NewQs("a", "b2", "c", "")
	set := SetQueryParams("https://example.com/path?a=b&c=d", qs)
	assert.Equal(t, "https://example.com/path?a=b2", set)
}

func TestGetQueryParam(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	fctx.Request.SetRequestURI("http://localhost/?foo=bar&baz=qux")
	c := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(c)

	ctx := &RequestContext{Fiber: c}

	result := GetQueryParam(ctx, "foo")
	assert.Equal(t, "bar", result)

	result = GetQueryParam(ctx, "baz")
	assert.Equal(t, "qux", result)

	result = GetQueryParam(ctx, "missing")
	assert.Equal(t, "", result)

	ctx.currentBrowserUrl = "http://localhost/?current=value"

	result = GetQueryParam(ctx, "current")
	assert.Equal(t, "value", result)

	// url params should override browser url
	fctx2 := &fasthttp.RequestCtx{}
	fctx2.Request.SetRequestURI("http://localhost/?foo=override")
	c2 := app.AcquireCtx(fctx2)
	defer app.ReleaseCtx(c2)

	ctx2 := &RequestContext{Fiber: c2, currentBrowserUrl: "http://localhost/?current=value"}
	result = GetQueryParam(ctx2, "foo")
	assert.Equal(t, "override", result)
}
