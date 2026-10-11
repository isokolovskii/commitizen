package git_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isokolovskii/commitizen/internal/git"
	"github.com/isokolovskii/commitizen/internal/git/gittest"
)

const (
	helloFile    = "hello.txt"
	ideaPath     = "notes/idea.md"
	prettyBody   = "--pretty=%B"
	helloContent = "hi\n"
)

// gitLog reads the body of the latest commit so assertions can compare full messages.
func gitLog(t *testing.T, dir string) string {
	t.Helper()

	var stdout bytes.Buffer
	cmd := exec.CommandContext(t.Context(), "git", "log", "-1", prettyBody)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	require.NoError(t, cmd.Run())

	return stdout.String()
}

func TestRootReturnsWorkingTreeRoot(t *testing.T) {
	t.Parallel()

	repo := gittest.Repo(t)

	root, err := git.NewRunner(repo).Root()

	require.NoError(t, err)
	want, evalErr := filepath.EvalSymlinks(repo)
	require.NoError(t, evalErr)
	assert.Equal(t, want, root)
}

func TestRootOutsideRepository(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	_, err := git.NewRunner(dir).Root()

	require.ErrorIs(t, err, git.ErrNotRepository)
}

func TestStagedFilesListsStagedPaths(t *testing.T) {
	t.Parallel()

	repo := gittest.Repo(t)
	gittest.WriteFile(t, repo, helloFile, helloContent)
	gittest.WriteFile(t, repo, ideaPath, "# idea\n")

	runner := git.NewRunner(repo)

	staged, err := runner.StagedFiles()
	require.NoError(t, err)
	assert.Empty(t, staged)

	gittest.Stage(t, repo, helloFile)
	gittest.Stage(t, repo, ideaPath)

	staged, err = runner.StagedFiles()

	require.NoError(t, err)
	assert.Equal(t, []string{helloFile, ideaPath}, staged)
}

func TestCommitCreatesCommitWithMessage(t *testing.T) {
	t.Parallel()

	repo := gittest.Repo(t)
	gittest.WriteFile(t, repo, helloFile, helloContent)
	gittest.Stage(t, repo, helloFile)

	runner := git.NewRunner(repo)

	require.NoError(t, runner.Commit("feat: initial\n"))

	logged := gitLog(t, repo)
	assert.Equal(t, "feat: initial", strings.TrimRight(logged, "\n"))

	staged, err := runner.StagedFiles()

	require.NoError(t, err)
	assert.Empty(t, staged)
}

func TestCommitMultiLineMessage(t *testing.T) {
	t.Parallel()

	repo := gittest.Repo(t)
	gittest.WriteFile(t, repo, helloFile, helloContent)
	gittest.Stage(t, repo, helloFile)

	message := "feat(api)!: add endpoint\n\nAdds the endpoint.\n\nBREAKING CHANGE: payload changed\nIssue: TEST-1\n"

	require.NoError(t, git.NewRunner(repo).Commit(message))

	logged := gitLog(t, repo)
	assert.Equal(t, message, strings.TrimRight(logged, "\n")+"\n")
}

func TestCommitOutsideRepository(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	err := git.NewRunner(dir).Commit("feat: nope\n")

	require.ErrorIs(t, err, git.ErrNotRepository)
}
