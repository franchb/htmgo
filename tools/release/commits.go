package main

import (
	"regexp"
	"strings"
)

// Commit is one commit on the release branch.
type Commit struct {
	Hash    string
	Subject string
	Body    string
}

// autoUpdateSubject is the message of the commits that
// .github/workflows/update-framework-dep.yml pushes to bump the htmgo
// requirements of the other modules.
const autoUpdateSubject = "Auto-update HTMGO framework version"

// Kind is the release-notes section a commit belongs to.
type Kind int

const (
	KindIgnored Kind = iota
	KindBreaking
	KindFeature
	KindFix
	KindPerf
	KindSecurity
	KindRevert
	KindDeps
	KindAutoUpdate
)

var kindTitles = map[Kind]string{
	KindBreaking: "Breaking changes",
	KindFeature:  "Features",
	KindFix:      "Fixes",
	KindPerf:     "Performance",
	KindSecurity: "Security",
	KindRevert:   "Reverts",
	KindDeps:     "Dependencies",
}

// sectionOrder is the order of sections in release notes. KindAutoUpdate is
// summarised under Dependencies rather than listed.
var sectionOrder = []Kind{KindBreaking, KindFeature, KindFix, KindPerf, KindSecurity, KindRevert, KindDeps}

// headerRe matches a Conventional Commits header: type(scope)!: description.
// https://www.conventionalcommits.org/en/v1.0.0/#specification
var headerRe = regexp.MustCompile(`^([a-zA-Z]+)(?:\(([^()]*)\))?(!)?: \S`)

// breakingFooterRe matches the BREAKING CHANGE footer (spec items 11-12, 16).
var breakingFooterRe = regexp.MustCompile(`(?m)^BREAKING[ -]CHANGE: \S`)

// Classify maps a commit to its release-notes section and the version bump it
// asks for:
//
//	type! or a BREAKING CHANGE footer    major
//	feat                                 minor
//	fix, perf, security, revert, deps    patch
//	build(deps), chore(deps-dev), ...    patch (Dependabot)
//	the dependency auto-update commit    patch
//	anything else (chore, ci, docs, ...) no release
func Classify(c Commit) (Kind, Bump) {
	if c.Subject == autoUpdateSubject {
		return KindAutoUpdate, BumpPatch
	}
	if strings.HasPrefix(c.Subject, `Revert "`) {
		return KindRevert, BumpPatch
	}
	m := headerRe.FindStringSubmatch(c.Subject)
	if m == nil {
		return KindIgnored, BumpNone
	}
	typ, scope, bang := strings.ToLower(m[1]), strings.ToLower(m[2]), m[3] == "!"
	if bang || breakingFooterRe.MatchString(c.Body) {
		return KindBreaking, BumpMajor
	}
	switch typ {
	case "feat":
		return KindFeature, BumpMinor
	case "fix":
		return KindFix, BumpPatch
	case "perf":
		return KindPerf, BumpPatch
	case "security":
		return KindSecurity, BumpPatch
	case "revert":
		return KindRevert, BumpPatch
	case "deps":
		return KindDeps, BumpPatch
	case "build", "chore":
		if scope == "deps" || strings.HasPrefix(scope, "deps-") {
			return KindDeps, BumpPatch
		}
	}
	return KindIgnored, BumpNone
}
