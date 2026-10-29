package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testRepo is a throwaway git repository.
type testRepo struct {
	t *testing.T
	Git
}

func newRepo(t *testing.T) *testRepo {
	t.Helper()
	r := &testRepo{t, Git{Dir: t.TempDir()}}
	r.git("init", "-q", "-b", "master")
	r.git("config", "user.name", "test")
	r.git("config", "user.email", "test@example.com")
	r.git("config", "commit.gpgsign", "false")
	r.git("config", "tag.gpgsign", "false")
	return r
}

func (r *testRepo) git(args ...string) string {
	r.t.Helper()
	out, err := r.Run(args...)
	if err != nil {
		r.t.Fatal(err)
	}
	return strings.TrimSpace(out)
}

func (r *testRepo) write(name, content string) {
	r.t.Helper()
	p := filepath.Join(r.Dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		r.t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		r.t.Fatal(err)
	}
}

// commit writes name (appending a line, so every commit changes it) and
// commits it with message.
func (r *testRepo) commit(name, message string) {
	r.t.Helper()
	p := filepath.Join(r.Dir, filepath.FromSlash(name))
	old, _ := os.ReadFile(p)
	r.write(name, string(old)+message+"\n")
	r.git("add", "-A")
	r.git("commit", "-q", "-m", message)
}

// baseline creates the layout of this repository in miniature:
//
//	lib/       example.com/r/lib/v2, tagged lib/v2.0.0
//	lib/inner  example.com/r/lib/inner, a nested module, never tagged
//	cli/       example.com/r/cli/v2, tagged cli/v2.0.0, builds the binary
//	app/       example.com/r/app, never tagged (like the examples)
//
// plus the repository tag v2.0.0.
func baseline(t *testing.T) *testRepo {
	r := newRepo(t)
	r.write("lib/go.mod", "module example.com/r/lib/v2\n\ngo 1.27.0\n")
	r.write("lib/inner/go.mod", "module example.com/r/lib/inner\n\ngo 1.27.0\n")
	r.write("cli/go.mod", "module example.com/r/cli/v2\n\ngo 1.27.0\n\nrequire example.com/r/lib/v2 v2.0.1-0.20260101000000-abcdefabcdef\n")
	r.write("app/go.mod", "module example.com/r/app\n\ngo 1.27.0\n")
	r.commit("README", "chore: initial")
	r.git("tag", "lib/v2.0.0")
	r.git("tag", "cli/v2.0.0")
	r.git("tag", "v2.0.0")
	// Tags that are not baselines: old unprefixed and pre-v2 versions, and a
	// prerelease above the latest release.
	r.git("tag", "lib/1.0.4")
	r.git("tag", "lib/v1.5.0")
	r.git("tag", "lib/v2.1.0-rc.1")
	return r
}

func plan(t *testing.T, r *testRepo) *Plan {
	t.Helper()
	p, err := MakePlan(r.Git, Config{Head: "HEAD", RepoURL: "https://github.com/o/r", BinaryDir: "cli", BinaryName: "tool"})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func tags(p *Plan) []string {
	var ts []string
	for _, m := range p.Modules {
		ts = append(ts, m.Tag)
	}
	if p.Root != nil {
		ts = append(ts, p.Root.Tag)
	}
	return ts
}

func assertTags(t *testing.T, p *Plan, want ...string) {
	t.Helper()
	if got := tags(p); strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("tags = %v, want %v", got, want)
	}
}

func TestPlanFeature(t *testing.T) {
	t.Parallel()
	r := baseline(t)
	r.commit("lib/a.go", "fix(lib): a")
	r.commit("lib/b.go", "feat(lib): b (#2)")
	r.commit("app/main.go", "feat(app): never released")
	p := plan(t, r)
	assertTags(t, p, "lib/v2.1.0", "v2.1.0")
	if p.Binary == nil || p.Binary.Tag != "cli/v2.0.0" || p.Binary.New {
		t.Errorf("binary = %+v, want the existing cli/v2.0.0", p.Binary)
	}
	notes := p.Modules[0].Notes
	for _, want := range []string{
		"go get example.com/r/lib/v2@v2.1.0",
		"### Features\n\n- feat(lib): b (#2)",
		"### Fixes\n\n- fix(lib): a",
		"https://github.com/o/r/compare/lib/v2.0.0...lib/v2.1.0",
	} {
		if !strings.Contains(notes, want) {
			t.Errorf("notes lack %q:\n%s", want, notes)
		}
	}
	if !strings.Contains(p.Root.Notes, "(unchanged in this release)") || !strings.Contains(p.Root.Notes, "--repo o/r") {
		t.Errorf("root notes:\n%s", p.Root.Notes)
	}
}

func TestPlanBinaryAndAutoUpdate(t *testing.T) {
	t.Parallel()
	r := baseline(t)
	r.commit("lib/a.go", "fix(lib): a")
	r.commit("cli/go.mod", autoUpdateSubject)
	r.commit("cli/go.mod", autoUpdateSubject)
	p := plan(t, r)
	assertTags(t, p, "cli/v2.0.1", "lib/v2.0.1", "v2.0.1")
	if !p.Binary.New || p.Binary.Tag != "cli/v2.0.1" || p.Binary.Version != "v2.0.1" {
		t.Errorf("binary = %+v", p.Binary)
	}
	notes := p.Modules[0].Notes
	for _, want := range []string{
		"go install example.com/r/cli/v2@v2.0.1",
		"(2 automatic commits): `example.com/r/lib/v2 v2.0.1-0.20260101000000-abcdefabcdef`",
	} {
		if !strings.Contains(notes, want) {
			t.Errorf("notes lack %q:\n%s", want, notes)
		}
	}
	if strings.Contains(notes, "go get") {
		t.Errorf("binary notes should only say go install:\n%s", notes)
	}
}

func TestPlanNothingToRelease(t *testing.T) {
	t.Parallel()
	r := baseline(t)
	r.commit("lib/a.go", "chore(lib): tidy")
	r.commit("lib/inner/a.go", "feat(inner): nested modules are separate")
	r.commit("docs.md", "feat: outside every module")
	p := plan(t, r)
	assertTags(t, p)
	if p.Root != nil {
		t.Errorf("root = %+v, want nil", p.Root)
	}
}

func TestPlanContinuesFromLatestTag(t *testing.T) {
	t.Parallel()
	r := baseline(t)
	r.commit("lib/a.go", "feat(lib): a")
	r.git("tag", "lib/v2.1.0")
	r.git("tag", "v2.1.0")
	r.commit("lib/b.go", "fix(lib): b")
	assertTags(t, plan(t, r), "lib/v2.1.1", "v2.1.1")
}

func TestPlanIgnoresTagsOffHistory(t *testing.T) {
	t.Parallel()
	r := baseline(t)
	r.git("checkout", "-q", "-b", "side")
	r.commit("lib/side.go", "feat(lib): side")
	r.git("tag", "lib/v2.5.0")
	r.git("checkout", "-q", "master")
	r.commit("lib/a.go", "fix(lib): a")
	assertTags(t, plan(t, r), "lib/v2.0.1", "v2.0.1")

	// A collision with a tag off history is an error, not an overwrite.
	r.git("tag", "lib/v2.0.1", "side")
	if _, err := MakePlan(r.Git, Config{Head: "HEAD"}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("err = %v, want tag collision", err)
	}
}

func TestPlanBreakingChangeNeedsNewModulePath(t *testing.T) {
	t.Parallel()
	r := baseline(t)
	r.commit("lib/a.go", "feat(lib)!: remove Foo")
	_, err := MakePlan(r.Git, Config{Head: "HEAD"})
	if err == nil || !strings.Contains(err.Error(), "move the module path to /v3") || !strings.Contains(err.Error(), "remove Foo") {
		t.Fatalf("err = %v, want breaking-change error", err)
	}

	// Moving the module path to /v3 releases v3.0.0; the repository tag
	// moves by a minor version only.
	r.write("lib/go.mod", "module example.com/r/lib/v3\n\ngo 1.27.0\n")
	r.git("commit", "-q", "-am", "feat(lib)!: module path /v3")
	p := plan(t, r)
	assertTags(t, p, "lib/v3.0.0", "v2.1.0")
	if p.Modules[0].Bump != "major" || !strings.Contains(p.Modules[0].Notes, "### Breaking changes") {
		t.Errorf("release = %+v\n%s", p.Modules[0], p.Modules[0].Notes)
	}
}

func TestPlanRejectsReplace(t *testing.T) {
	t.Parallel()
	r := baseline(t)
	r.write("lib/go.mod", "module example.com/r/lib/v2\n\ngo 1.27.0\n\nreplace example.com/x => ../x\n")
	r.git("commit", "-q", "-am", "fix(lib): local replace")
	if _, err := MakePlan(r.Git, Config{Head: "HEAD"}); err == nil || !strings.Contains(err.Error(), "replace") {
		t.Fatalf("err = %v, want replace error", err)
	}
}

func TestParseGoMod(t *testing.T) {
	t.Parallel()
	m, err := parseGoMod(`module github.com/o/r/framework/v2 // comment

go 1.27.0

require github.com/o/r/tools/x/v2 v2.1.1

require (
	github.com/a/b v1.0.0
	github.com/c/d v1.2.3 // indirect
)
`)
	if err != nil {
		t.Fatal(err)
	}
	if m.Path != "github.com/o/r/framework/v2" || m.Major != 2 || m.Replace || len(m.Requires) != 3 || m.Requires[2] != (Requirement{"github.com/c/d", "v1.2.3"}) {
		t.Errorf("parseGoMod = %+v", m)
	}
	if m, _ := parseGoMod("module example.com/x\n"); m.Major != 1 {
		t.Errorf("major without suffix = %d, want 1", m.Major)
	}
	if m, _ := parseGoMod("module example.com/x\nreplace (\n\ta => ../a\n)\n"); !m.Replace {
		t.Error("replace block not detected")
	}
}
