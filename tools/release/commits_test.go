package main

import "testing"

func TestClassify(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		subject, body string
		kind          Kind
		bump          Bump
	}{
		{"feat(framework): h.Validation (#83)", "", KindFeature, BumpMinor},
		{"feat: thing", "", KindFeature, BumpMinor},
		{"fix(cli): stop the old server (#81)", "", KindFix, BumpPatch},
		{"perf(h): faster render", "", KindPerf, BumpPatch},
		{"security: bump golang.org/x/net (#65)", "", KindSecurity, BumpPatch},
		{"deps(framework/assets/js): bump undici (#86)", "", KindDeps, BumpPatch},
		{"deps: bump fiber (#69)", "", KindDeps, BumpPatch},
		{"build(deps): bump esbuild (#61)", "", KindDeps, BumpPatch},
		{"build(deps-dev): bump undici (#56)", "", KindDeps, BumpPatch},
		{"revert: feat: thing", "", KindRevert, BumpPatch},
		{`Revert "feat: thing"`, "", KindRevert, BumpPatch},
		{autoUpdateSubject, "", KindAutoUpdate, BumpPatch},
		{"feat!: drop Go 1.26", "", KindBreaking, BumpMajor},
		{"fix(h)!: rename Foo", "", KindBreaking, BumpMajor},
		{"refactor: move code", "BREAKING CHANGE: Foo is gone", KindBreaking, BumpMajor},
		{"fix: x", "Details.\n\nBREAKING-CHANGE: y", KindBreaking, BumpMajor},
		// Not breaking: the footer must start a line and be spelled exactly.
		{"fix: x", "mentions BREAKING CHANGE: inline", KindFix, BumpPatch},
		{"fix: x", "BREAKING: old-style footer", KindFix, BumpPatch},
		{"chore: staticcheck fixes (#68)", "", KindIgnored, BumpNone},
		{"ci(update-htmgo-dep): retry go get (#87)", "", KindIgnored, BumpNone},
		{"docs: typo", "", KindIgnored, BumpNone},
		{"test: fix parser tests (#66)", "", KindIgnored, BumpNone},
		{"build: tsup config", "", KindIgnored, BumpNone},
		{"refactor: tidy", "", KindIgnored, BumpNone},
		{"Merge branch 'master'", "", KindIgnored, BumpNone},
		{"feat:no space", "", KindIgnored, BumpNone},
		{"Update README.md", "", KindIgnored, BumpNone},
	} {
		kind, bump := Classify(Commit{Subject: tc.subject, Body: tc.body})
		if kind != tc.kind || bump != tc.bump {
			t.Errorf("Classify(%q, %q) = %v, %v; want %v, %v", tc.subject, tc.body, kind, bump, tc.kind, tc.bump)
		}
	}
}
