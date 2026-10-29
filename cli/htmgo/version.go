package main

import "runtime/debug"

// version is set at link time by the release workflow
// (-ldflags "-X main.version=v2.3.4"). It is empty in other builds.
var version string

// cliVersion reports the release version. Release binaries carry it in
// version; `go install github.com/franchb/htmgo/cli/htmgo/v2@vX.Y.Z` records it
// in the build info instead. A build from a source checkout reports "dev".
func cliVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
