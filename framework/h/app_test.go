package h

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

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
