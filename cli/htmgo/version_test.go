package main

import "testing"

func TestCliVersionPrefersLinkTimeVersion(t *testing.T) {
	old := version
	t.Cleanup(func() { version = old })

	version = "v2.3.4"
	if got := cliVersion(); got != "v2.3.4" {
		t.Fatalf("cliVersion() = %q, want v2.3.4", got)
	}
}
