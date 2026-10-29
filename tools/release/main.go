// Command release plans the next release of every Go module in the
// repository from Conventional Commits and the existing per-module tags.
//
//	go run . -repo ../.. -head master -out /tmp/release
//
// It writes plan.json and one notes/<tag>.md per release into -out and
// prints a summary. It never creates tags; the release workflow does that
// from plan.json. See CONTRIBUTING.md ("Releases") for the rules.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	repo := flag.String("repo", ".", "repository root")
	head := flag.String("head", "HEAD", "commit to release")
	out := flag.String("out", "", "directory for plan.json and notes/ (required)")
	repoURL := flag.String("repo-url", "", "repository URL for links in release notes, e.g. https://github.com/owner/repo")
	binary := flag.String("binary", "", "module that builds the released binary, as dir=name (e.g. cli/htmgo=htmgo)")
	flag.Parse()
	if *out == "" {
		fmt.Fprintln(os.Stderr, "release: -out is required")
		os.Exit(2)
	}
	cfg := Config{Head: *head, RepoURL: strings.TrimSuffix(*repoURL, "/")}
	if *binary != "" {
		dir, name, ok := strings.Cut(*binary, "=")
		if !ok || dir == "" || name == "" {
			fmt.Fprintln(os.Stderr, "release: -binary must be dir=name")
			os.Exit(2)
		}
		cfg.BinaryDir, cfg.BinaryName = dir, name
	}
	plan, err := MakePlan(Git{Dir: *repo}, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "release:", err)
		os.Exit(1)
	}
	if err := write(*out, plan); err != nil {
		fmt.Fprintln(os.Stderr, "release:", err)
		os.Exit(1)
	}
	if err := summarize(os.Stdout, plan); err != nil {
		fmt.Fprintln(os.Stderr, "release:", err)
		os.Exit(1)
	}
}

// NotesFile is where the notes for tag are written, relative to -out.
func NotesFile(tag string) string {
	return filepath.Join("notes", strings.ReplaceAll(tag, "/", "_")+".md")
}

func write(dir string, p *Plan) error {
	if err := os.MkdirAll(filepath.Join(dir, "notes"), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "plan.json"), append(data, '\n'), 0o644); err != nil {
		return err
	}
	for _, r := range p.Modules {
		if err := os.WriteFile(filepath.Join(dir, NotesFile(r.Tag)), []byte(r.Notes), 0o644); err != nil {
			return err
		}
	}
	if p.Root != nil {
		return os.WriteFile(filepath.Join(dir, NotesFile(p.Root.Tag)), []byte(p.Root.Notes), 0o644)
	}
	return nil
}

func summarize(w io.Writer, p *Plan) error {
	var sb strings.Builder
	if p.Root == nil {
		fmt.Fprintf(&sb, "Nothing to release at %s.\n", short(p.Head))
	} else {
		fmt.Fprintf(&sb, "Release at %s:\n", short(p.Head))
		for _, r := range p.Modules {
			fmt.Fprintf(&sb, "  %-36s %-6s (from %s)\n", r.Tag, r.Bump, r.Previous)
		}
		fmt.Fprintf(&sb, "  %-36s %-6s (from %s)\n", p.Root.Tag, p.Root.Bump, p.Root.Previous)
		if b := p.Binary; b != nil {
			fmt.Fprintf(&sb, "Binary %s from %s\n", b.Name, b.Tag)
		}
	}
	_, err := io.WriteString(w, sb.String())
	return err
}
