package h

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseExtensions(t *testing.T) {
	// Restore env cache after test completes so we don't leak state.
	t.Cleanup(resetEnvCache)

	// Test when not in development
	t.Setenv("ENV", "")
	resetEnvCache()
	result := BaseExtensions()
	expected := "path-deps, response-targets, mutation-error, htmgo, sse"
	assert.Equal(t, expected, result)

	// Test when in development
	t.Setenv("ENV", "development")
	resetEnvCache()
	result = BaseExtensions()
	expected = "path-deps, response-targets, mutation-error, htmgo, sse, livereload"
	assert.Equal(t, expected, result)
}
