package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Git runs git commands in one repository.
type Git struct{ Dir string }

func (g Git) Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.Dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func (g Git) Lines(args ...string) ([]string, error) {
	out, err := g.Run(args...)
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, l := range strings.Split(out, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	return lines, nil
}

// Log returns the non-merge commits in from..to (all of to's history when
// from is empty) that touch the given pathspecs, newest first.
func (g Git) Log(from, to string, pathspecs ...string) ([]Commit, error) {
	rev := to
	if from != "" {
		rev = from + ".." + to
	}
	args := append([]string{"log", "--no-merges", "--format=%H%x1f%s%x1f%b%x1e", rev, "--"}, pathspecs...)
	out, err := g.Run(args...)
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for _, rec := range strings.Split(out, "\x1e") {
		rec = strings.TrimLeft(rec, "\n")
		if rec == "" {
			continue
		}
		f := strings.SplitN(rec, "\x1f", 3)
		if len(f) != 3 {
			return nil, fmt.Errorf("git log: malformed record %q", rec)
		}
		commits = append(commits, Commit{Hash: f[0], Subject: f[1], Body: strings.TrimSpace(f[2])})
	}
	return commits, nil
}
