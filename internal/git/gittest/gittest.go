// Package gittest provides helpers that drive the real git binary inside
// temporary repositories for integration tests.
package gittest

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	gitBinary     = "git"
	configCommand = "config"
	dirPerm       = 0o750
	filePerm      = 0o600
)

// Repo initializes a fresh git repository in a new temporary directory and
// returns its path. The repository has a pinned identity so commits are deterministic.
func Repo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	Git(t, dir, "init")
	Git(t, dir, configCommand, "user.name", "Test User")
	Git(t, dir, configCommand, "user.email", "test@example.com")

	return dir
}

// Git runs a git command in dir with a pinned environment, failing the test on error.
func Git(t *testing.T, dir string, args ...string) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(t.Context(), gitBinary, args...)
	cmd.Dir = dir
	cmd.Env = env(t)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, stderr.String())
	}
}

// WriteFile writes content to a file under dir, creating parent directories as needed.
func WriteFile(t *testing.T, dir, name, content string) {
	t.Helper()

	path := filepath.Join(dir, name)
	parent := filepath.Dir(path)

	err := os.MkdirAll(parent, dirPerm)
	if err != nil {
		t.Fatalf("creating %s: %v", parent, err)
	}

	err = os.WriteFile(path, []byte(content), filePerm)
	if err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// Stage adds path to the repository index.
func Stage(t *testing.T, dir, path string) {
	t.Helper()

	Git(t, dir, "add", path)
}

// env returns the process environment pinned to an isolated home directory so
// user-level git configuration cannot leak into tests.
func env(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()

	return append(os.Environ(),
		"HOME="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
}
