package main

import (
	"bufio"
	"errors"
	"fmt"
	"path"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// Module is a Go module in the repository. Dir is its slash-separated path
// from the repository root; its tags are "<Dir>/vX.Y.Z".
type Module struct {
	Dir      string
	Path     string // module path from go.mod
	Major    int    // major version the module path allows (/vN suffix; 1 without one)
	Replace  bool   // go.mod has replace directives
	Requires []Requirement
}

type Requirement struct{ Path, Version string }

// ModuleRelease is the next release of one module.
type ModuleRelease struct {
	Dir      string    `json:"dir"`
	Module   string    `json:"module"`
	Previous string    `json:"previous"` // previous tag
	Tag      string    `json:"tag"`
	Version  string    `json:"version"`
	Bump     string    `json:"bump"`
	Notes    string    `json:"-"`
	commits  []classed // newest first
}

// RootRelease is the repository-wide release ("vX.Y.Z"). It carries the
// binaries and is the release GitHub shows as latest.
type RootRelease struct {
	Previous string `json:"previous"`
	Tag      string `json:"tag"`
	Bump     string `json:"bump"`
	Notes    string `json:"-"`
}

// Binary says where to build a released program from: the module's new tag
// when this plan releases it, else its latest tag.
type Binary struct {
	Dir     string `json:"dir"`
	Name    string `json:"name"`
	Module  string `json:"module"`
	Tag     string `json:"tag"`
	Version string `json:"version"`
	New     bool   `json:"new"`
}

type Plan struct {
	Head    string          `json:"head"`
	Root    *RootRelease    `json:"root"`
	Modules []ModuleRelease `json:"modules"`
	Binary  *Binary         `json:"binary,omitempty"`
}

type classed struct {
	Commit
	Kind Kind
}

type Config struct {
	Head       string // commit to release
	RepoURL    string // https://github.com/owner/repo, for links in notes
	BinaryDir  string // module that builds the released binary, e.g. cli/htmgo
	BinaryName string
}

var moduleLineRe = regexp.MustCompile(`/v([2-9]|[1-9][0-9]+)$`)

// LoadModules reads every go.mod tracked at head.
func LoadModules(g Git, head string) ([]Module, error) {
	files, err := g.Lines("ls-tree", "-r", "--name-only", head)
	if err != nil {
		return nil, err
	}
	var mods []Module
	for _, f := range files {
		if path.Base(f) != "go.mod" {
			continue
		}
		data, err := g.Run("show", head+":"+f)
		if err != nil {
			return nil, err
		}
		m, err := parseGoMod(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		m.Dir = path.Dir(f)
		mods = append(mods, m)
	}
	return mods, nil
}

// parseGoMod reads the few go.mod facts the planner needs. It is not a full
// parser (golang.org/x/mod/modfile is), but go.mod files written by the go
// command always use these line shapes.
func parseGoMod(data string) (Module, error) {
	var m Module
	block := ""
	sc := bufio.NewScanner(strings.NewReader(data))
	for sc.Scan() {
		line, _, _ := strings.Cut(sc.Text(), "//")
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if block != "" {
			if f[0] == ")" {
				block = ""
				continue
			}
			f = append([]string{block}, f...)
		} else if len(f) == 2 && f[1] == "(" {
			block = f[0]
			continue
		}
		switch f[0] {
		case "module":
			if len(f) >= 2 {
				m.Path = strings.Trim(f[1], `"`)
			}
		case "replace":
			m.Replace = true
		case "require":
			if len(f) >= 3 {
				m.Requires = append(m.Requires, Requirement{Path: f[1], Version: f[2]})
			}
		}
	}
	if err := sc.Err(); err != nil {
		return m, err
	}
	if m.Path == "" {
		return m, errors.New("no module directive")
	}
	m.Major = 1
	if sm := moduleLineRe.FindStringSubmatch(m.Path); sm != nil {
		m.Major, _ = strconv.Atoi(sm[1])
	}
	return m, nil
}

// latestTag returns the highest stable tag "<prefix>vX.Y.Z" whose major is
// accepted by majorOK.
func latestTag(tags []string, prefix string, majorOK func(int) bool) (string, Version, bool) {
	var best Version
	bestTag, found := "", false
	for _, t := range tags {
		rest, ok := strings.CutPrefix(t, prefix)
		if !ok {
			continue
		}
		v, err := ParseVersion(rest)
		if err != nil || v.Pre != "" || !majorOK(v.Major) {
			continue
		}
		if !found || v.Compare(best) > 0 {
			best, bestTag, found = v, t, true
		}
	}
	return bestTag, best, found
}

// majorAccepts reports whether a module whose path allows major m can carry
// a tag with major v. Paths without a /vN suffix carry v0 and v1.
func majorAccepts(m, v int) bool {
	if m == 1 {
		return v <= 1
	}
	return v == m
}

// MakePlan works out the next release of every released module.
//
// A module is released when it has at least one tag "<dir>/vX.Y.Z"; others
// (examples, the site, templates) are never tagged. Its next version comes
// from the commits since its latest tag that touch its directory, excluding
// modules nested inside it. A breaking change is only released when the
// module path already names the new major version (Go semantic import
// versioning: https://go.dev/ref/mod#major-version-suffixes); otherwise
// planning fails, because tagging v3.0.0 on a /v2 module path is invalid and
// shipping the break as v2.x would break consumers on `go get -u`.
func MakePlan(g Git, cfg Config) (*Plan, error) {
	head, err := g.Run("rev-parse", "--verify", cfg.Head+"^{commit}")
	if err != nil {
		return nil, err
	}
	head = strings.TrimSpace(head)
	mods, err := LoadModules(g, head)
	if err != nil {
		return nil, err
	}
	tags, err := g.Lines("tag", "--list")
	if err != nil {
		return nil, err
	}
	// A release must not reuse a tag that is not on head's history.
	merged, err := g.Lines("tag", "--merged", head)
	if err != nil {
		return nil, err
	}

	plan := &Plan{Head: head}
	var errs []error
	maxBump := BumpNone
	for _, m := range mods {
		prefix := m.Dir + "/"
		if _, _, any := latestTag(tags, prefix, func(int) bool { return true }); !any {
			continue // never released
		}
		prevTag, prev, ok := latestTag(merged, prefix, func(v int) bool { return majorAccepts(m.Major, v) })
		newMajor := false
		if !ok {
			// The module path moved to a new major (/v2 -> /v3): continue from
			// the previous major's latest tag.
			prevTag, prev, ok = latestTag(merged, prefix, func(v int) bool { return majorAccepts(m.Major-1, v) })
			if !ok {
				errs = append(errs, fmt.Errorf("%s: no tag %sv%d.x.y (or v%d.x.y) reachable from %s; tag it by hand first", m.Dir, prefix, m.Major, m.Major-1, short(head)))
				continue
			}
			newMajor = true
		}

		specs := []string{m.Dir}
		for _, n := range mods {
			if n.Dir != m.Dir && strings.HasPrefix(n.Dir, prefix) {
				specs = append(specs, ":(exclude)"+n.Dir)
			}
		}
		commits, err := g.Log(prevTag, head, specs...)
		if err != nil {
			return nil, err
		}
		bump := BumpNone
		var cs []classed
		for _, c := range commits {
			k, b := Classify(c)
			if k == KindIgnored {
				continue
			}
			cs = append(cs, classed{c, k})
			bump = max(bump, b)
		}
		if bump == BumpNone {
			continue
		}

		var next Version
		switch {
		case newMajor:
			next = Version{Major: m.Major}
		case bump == BumpMajor:
			var breaking []string
			for _, c := range cs {
				if c.Kind == KindBreaking {
					breaking = append(breaking, fmt.Sprintf("%s %s", short(c.Hash), c.Subject))
				}
			}
			errs = append(errs, fmt.Errorf("%s: breaking changes since %s, but the module path %s still names v%d. "+
				"Either move the module path to /v%d (a release then tags v%d.0.0), or tag the release by hand after deciding it is compatible:\n  %s",
				m.Dir, prevTag, m.Path, m.Major, m.Major+1, m.Major+1, strings.Join(breaking, "\n  ")))
			continue
		default:
			next = prev.Bump(bump)
		}
		if newMajor {
			bump = BumpMajor
		}
		if m.Replace {
			errs = append(errs, fmt.Errorf("%s: go.mod has replace directives; a published module must not (go install rejects them)", m.Dir))
			continue
		}
		tag := prefix + next.String()
		if slices.Contains(tags, tag) {
			errs = append(errs, fmt.Errorf("%s: tag %s already exists but is not reachable from %s", m.Dir, tag, short(head)))
			continue
		}
		r := ModuleRelease{
			Dir: m.Dir, Module: m.Path, Previous: prevTag, Tag: tag,
			Version: next.String(), Bump: bump.String(), commits: cs,
		}
		r.Notes = moduleNotes(cfg, m, r)
		plan.Modules = append(plan.Modules, r)
		maxBump = max(maxBump, bump)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	sort.Slice(plan.Modules, func(i, j int) bool { return plan.Modules[i].Dir < plan.Modules[j].Dir })

	if len(plan.Modules) > 0 {
		prevTag, prev, ok := latestTag(merged, "", func(int) bool { return true })
		if !ok {
			return nil, fmt.Errorf("no repository tag vX.Y.Z reachable from %s; tag it by hand first", short(head))
		}
		// The repository tag names no Go module, so a module's new major is
		// only a minor release of the whole.
		rootBump := min(maxBump, BumpMinor)
		next := prev.Bump(rootBump)
		if slices.Contains(tags, next.String()) {
			return nil, fmt.Errorf("tag %s already exists but is not reachable from %s", next, short(head))
		}
		plan.Root = &RootRelease{Previous: prevTag, Tag: next.String(), Bump: rootBump.String()}
	}

	if cfg.BinaryDir != "" {
		b, err := binaryFor(cfg, mods, merged, plan)
		if err != nil {
			return nil, err
		}
		plan.Binary = b
	}
	if plan.Root != nil {
		plan.Root.Notes = rootNotes(cfg, plan)
	}
	return plan, nil
}

func binaryFor(cfg Config, mods []Module, merged []string, plan *Plan) (*Binary, error) {
	i := slices.IndexFunc(mods, func(m Module) bool { return m.Dir == cfg.BinaryDir })
	if i < 0 {
		return nil, fmt.Errorf("binary module %s: no go.mod", cfg.BinaryDir)
	}
	m := mods[i]
	b := &Binary{Dir: m.Dir, Name: cfg.BinaryName, Module: m.Path}
	for _, r := range plan.Modules {
		if r.Dir == m.Dir {
			b.Tag, b.Version, b.New = r.Tag, r.Version, true
			return b, nil
		}
	}
	tag, v, ok := latestTag(merged, m.Dir+"/", func(v int) bool { return majorAccepts(m.Major, v) })
	if !ok {
		return nil, fmt.Errorf("binary module %s: no tag reachable", m.Dir)
	}
	b.Tag, b.Version = tag, v.String()
	return b, nil
}

func short(h string) string {
	if len(h) > 7 {
		return h[:7]
	}
	return h
}

func moduleNotes(cfg Config, m Module, r ModuleRelease) string {
	var sb strings.Builder
	if r.Dir == cfg.BinaryDir {
		fmt.Fprintf(&sb, "```sh\ngo install %s@%s\n```\n\nPrebuilt `%s` binaries are attached to this release.\n", m.Path, r.Version, cfg.BinaryName)
	} else {
		fmt.Fprintf(&sb, "```sh\ngo get %s@%s\n```\n", m.Path, r.Version)
	}
	byKind := map[Kind][]classed{}
	autoUpdates := 0
	for _, c := range r.commits {
		if c.Kind == KindAutoUpdate {
			autoUpdates++
			continue
		}
		byKind[c.Kind] = append(byKind[c.Kind], c)
	}
	for _, k := range sectionOrder {
		cs := byKind[k]
		if len(cs) == 0 && (k != KindDeps || autoUpdates == 0) {
			continue
		}
		fmt.Fprintf(&sb, "\n### %s\n\n", kindTitles[k])
		for _, c := range cs {
			fmt.Fprintf(&sb, "- %s (%s)\n", c.Subject, short(c.Hash))
		}
		if k == KindDeps && autoUpdates > 0 {
			var reqs []string
			for _, q := range m.Requires {
				if strings.HasPrefix(q.Path, repoModulePrefix(m)) {
					reqs = append(reqs, fmt.Sprintf("`%s %s`", q.Path, q.Version))
				}
			}
			fmt.Fprintf(&sb, "- htmgo module requirements updated to master (%d automatic commit%s)", autoUpdates, plural(autoUpdates))
			if len(reqs) > 0 {
				fmt.Fprintf(&sb, ": %s", strings.Join(reqs, ", "))
			}
			sb.WriteString("\n")
		}
	}
	if cfg.RepoURL != "" {
		fmt.Fprintf(&sb, "\n**Full diff:** %s/compare/%s...%s\n", cfg.RepoURL, r.Previous, r.Tag)
	}
	return sb.String()
}

// repoModulePrefix returns the module path prefix shared by the
// repository's modules: "github.com/franchb/htmgo/" for module
// "github.com/franchb/htmgo/framework/v2" in directory "framework".
func repoModulePrefix(m Module) string {
	p := moduleLineRe.ReplaceAllString(m.Path, "")
	if prefix, ok := strings.CutSuffix(p, "/"+m.Dir); ok {
		return prefix + "/"
	}
	return p + "/"
}

func rootNotes(cfg Config, p *Plan) string {
	var sb strings.Builder
	sb.WriteString("| Module | Version | Previous |\n|---|---|---|\n")
	for _, r := range p.Modules {
		fmt.Fprintf(&sb, "| `%s` | %s | %s |\n", r.Module, releaseLink(cfg, r.Tag, r.Version), r.Previous)
	}
	if b := p.Binary; b != nil {
		fmt.Fprintf(&sb, "\nThe attached `%s` binaries are `%s` %s", b.Name, b.Module, b.Version)
		if !b.New {
			sb.WriteString(" (unchanged in this release)")
		}
		sb.WriteString(". Check a download with `sha256sum -c SHA256SUMS --ignore-missing`")
		if repo := ownerRepo(cfg.RepoURL); repo != "" {
			fmt.Fprintf(&sb, ", and its build provenance with `gh attestation verify <file> --repo %s`", repo)
		}
		sb.WriteString(".\n")
	}
	for _, r := range p.Modules {
		fmt.Fprintf(&sb, "\n## %s %s\n\n%s", r.Dir, r.Version, demoteHeadings(r.Notes))
	}
	if cfg.RepoURL != "" {
		fmt.Fprintf(&sb, "\n**Full diff:** %s/compare/%s...%s\n", cfg.RepoURL, p.Root.Previous, p.Root.Tag)
	}
	return sb.String()
}

// ownerRepo returns "owner/repo" for "https://github.com/owner/repo".
func ownerRepo(url string) string {
	parts := strings.Split(strings.TrimSuffix(url, "/"), "/")
	if len(parts) < 2 || url == "" {
		return ""
	}
	return strings.Join(parts[len(parts)-2:], "/")
}

func releaseLink(cfg Config, tag, text string) string {
	if cfg.RepoURL == "" {
		return text
	}
	return fmt.Sprintf("[%s](%s/releases/tag/%s)", text, cfg.RepoURL, tag)
}

func demoteHeadings(s string) string {
	return strings.ReplaceAll(s, "\n### ", "\n#### ")
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
