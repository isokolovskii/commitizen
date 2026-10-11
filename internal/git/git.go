// Package git abstracts the git operations the commit wizard needs.
package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type (
	// Runner abstracts the git operations required to inspect a repository
	// and create commits.
	Runner interface {
		// Root returns the absolute path of the repository working tree root.
		Root() (string, error)
		// StagedFiles returns the paths currently staged for commit.
		StagedFiles() ([]string, error)
		// Commit creates a commit with the given message.
		Commit(message string) error
	}

	// ExecRunner runs the real git binary through os/exec from a fixed working directory.
	ExecRunner struct {
		// Dir is the working directory every git invocation runs in.
		dir string
	}
)

const (
	gitBinary          = "git"
	notRepositoryMark  = "not a git repository"
	rootFlag           = "--show-toplevel"
	stagedFilesCommand = "diff"
	stagedFilesFlag    = "--cached"
	nameOnlyFlag       = "--name-only"
	commitFileFlag     = "--file=-"
)

// ErrNotRepository reports that the working directory is not inside a git repository.
var (
	ErrNotRepository = errors.New("not a git repository")

	_ Runner = (*ExecRunner)(nil)
)

// NewRunner returns an ExecRunner that runs git from dir.
func NewRunner(dir string) *ExecRunner {
	return &ExecRunner{dir: dir}
}

// Root returns the absolute path of the repository working tree root.
func (runner *ExecRunner) Root() (string, error) {
	out, err := runner.output("", "rev-parse", rootFlag)
	if err != nil {
		return "", fmt.Errorf("resolving repository root: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

// StagedFiles returns the paths currently staged for commit.
func (runner *ExecRunner) StagedFiles() ([]string, error) {
	out, err := runner.output("", stagedFilesCommand, stagedFilesFlag, nameOnlyFlag)
	if err != nil {
		return nil, fmt.Errorf("listing staged files: %w", err)
	}

	files := []string{}

	for line := range strings.SplitSeq(string(out), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

// Commit creates a commit with the given message.
func (runner *ExecRunner) Commit(message string) error {
	_, err := runner.output(message, "commit", commitFileFlag)
	if err != nil {
		return fmt.Errorf("creating commit: %w", err)
	}

	return nil
}

// output runs git with the given arguments, feeding stdin to the process when non-empty.
func (runner *ExecRunner) output(stdin string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(context.Background(), gitBinary, args...)
	cmd.Dir = runner.dir
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		if strings.Contains(stderr.String(), notRepositoryMark) {
			return nil, ErrNotRepository
		}

		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return out, nil
}
