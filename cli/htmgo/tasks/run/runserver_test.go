package run

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResetDir_RecreatesExistingDirEmpty(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "temp-build")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "old.exe"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := resetDir(dir, time.Second); err != nil {
		t.Fatalf("resetDir: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("dir should exist: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("dir should be empty, has %d entries", len(entries))
	}
}

func TestResetDir_CreatesMissingDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "__htmgo", "temp-build")
	if err := resetDir(dir, time.Second); err != nil {
		t.Fatalf("resetDir: %v", err)
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Fatalf("expected directory, got %v %v", fi, err)
	}
}

func TestFindBinary(t *testing.T) {
	dir := t.TempDir()
	if _, err := findBinary(dir); err == nil {
		t.Fatal("expected an error for an empty build dir")
	}
	want := filepath.Join(dir, "app.exe")
	if err := os.WriteFile(want, []byte("x"), 0755); err != nil {
		t.Fatal(err)
	}
	got, err := findBinary(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
