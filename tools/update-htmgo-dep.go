package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const frameworkRepo = "github.com/franchb/htmgo/framework/v2"
const htmlToHtmgoRepo = "github.com/franchb/htmgo/tools/html-to-htmgo/v2"

var depsToUpdate = []string{
	frameworkRepo,
	htmlToHtmgoRepo,
}

const githubAPIURL = "https://api.github.com/repos/franchb/htmgo/commits"

// Commit represents the structure of a commit object returned by the GitHub API.
type Commit struct {
	Sha string `json:"sha"`
}

func RunCommand(path string, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = path // Set the working directory to the current path.
	// Most apps (htmgo-site, templates, examples) are not in go.work; resolve
	// every module from its own go.mod.
	cmd.Env = append(os.Environ(), "GOWORK=off")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s in %s: %w", command, strings.Join(args, " "), path, err)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	// Get the latest commit hash from the remote repository.
	latestCommitHash, err := getLatestCommitHash()
	if err != nil {
		return fmt.Errorf("getting latest commit hash: %w", err)
	}

	// Start from the current directory.
	rootDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	var errs []error
	// Walk through directories.
	err = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if d.Name() == "node_modules" || d.Name() == ".git" {
			return filepath.SkipDir
		}

		// Check if the directory contains a go.mod file.
		goModPath := filepath.Join(path, "go.mod")
		if fileExists(goModPath) {
			for _, dep := range depsToUpdate {
				if err := updateDepToLatestVersion(dep, goModPath, latestCommitHash); err != nil {
					errs = append(errs, err)
				}
			}
		}

		return nil
	})
	if err != nil {
		errs = append(errs, fmt.Errorf("walking directories: %w", err))
	}

	return errors.Join(errs...)
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func updateDepToLatestVersion(dep string, goModPath string, latestCommitHash string) error {
	required, err := requiresDep(dep, goModPath)
	if err != nil || !required {
		return err
	}
	dir := filepath.Dir(goModPath)
	// Run go get github.com/franchb/htmgo/framework/v2@<latestCommitHash>.
	fmt.Printf("Updating %s to %s in %s\n", dep, latestCommitHash, dir)
	if err := withRetry(goGetRetryDelays, func() error {
		return RunCommand(dir, "go", "get", fmt.Sprintf("%s@%s", dep, latestCommitHash))
	}); err != nil {
		return err
	}
	return RunCommand(dir, "go", "mod", "tidy")
}

// goGetRetryDelays are the waits between go get attempts. This workflow runs
// right after a push to master, and the module proxy can briefly fail to
// resolve a nested module (e.g. framework/v2) at the new commit and fall back
// to the repository root module ("does not contain package"). A few minutes of
// backoff covers that lag.
var goGetRetryDelays = []time.Duration{
	15 * time.Second,
	30 * time.Second,
	time.Minute,
	2 * time.Minute,
}

// withRetry calls fn until it succeeds, sleeping delays[i] after the i-th
// failure. It returns the last error once the delays are exhausted.
func withRetry(delays []time.Duration, fn func() error) error {
	err := fn()
	for _, delay := range delays {
		if err == nil {
			return nil
		}
		fmt.Fprintf(os.Stderr, "%v; retrying in %s\n", err, delay)
		time.Sleep(delay)
		err = fn()
	}
	return err
}

// requiresDep reports whether the go.mod file has a require directive for dep.
// The module's own "module" line and "replace" lines do not count.
func requiresDep(dep string, goModPath string) (bool, error) {
	file, err := os.Open(goModPath)
	if err != nil {
		return false, fmt.Errorf("opening %s: %w", goModPath, err)
	}
	defer file.Close()

	inRequireBlock := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		switch {
		case len(fields) == 0:
		case inRequireBlock && fields[0] == ")":
			inRequireBlock = false
		case inRequireBlock && fields[0] == dep:
			return true, nil
		case fields[0] == "require" && len(fields) > 1 && fields[1] == "(":
			inRequireBlock = true
		case fields[0] == "require" && len(fields) > 1 && fields[1] == dep:
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("reading %s: %w", goModPath, err)
	}

	return false, nil
}

// getLatestCommitHash fetches the latest commit hash from the GitHub API for the framework repository.
func getLatestCommitHash() (string, error) {
	// Create an HTTP request to the GitHub API to get the latest commit.
	resp, err := http.Get(githubAPIURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest commit: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK.
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch latest commit: status code %d", resp.StatusCode)
	}

	// Parse the JSON response.
	var commits []Commit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "", fmt.Errorf("failed to decode JSON response: %v", err)
	}

	if len(commits) == 0 {
		return "", fmt.Errorf("no commits found in repository")
	}

	// Return the SHA of the latest commit.
	return commits[0].Sha, nil
}
