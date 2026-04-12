package h

import "os"

// Environment flags are cached at startup so that hot-path checks
// (e.g. inside middleware) do not call os.Getenv on every request.
var (
	isDev   = os.Getenv("ENV") == "development"
	isProd  = os.Getenv("ENV") == "production"
	isWatch = os.Getenv("WATCH_MODE") == "true"
)

func IsDevelopment() bool { return isDev }
func IsProduction() bool  { return isProd }
func IsWatchMode() bool   { return isWatch }

// resetEnvCache re-reads environment variables into the cached flags.
// This is intended for use in tests only (unexported, accessible from
// same-package test files).
func resetEnvCache() {
	isDev = os.Getenv("ENV") == "development"
	isProd = os.Getenv("ENV") == "production"
	isWatch = os.Getenv("WATCH_MODE") == "true"
}
