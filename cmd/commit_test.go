package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	stdconventional "github.com/isokolovskii/commitizen/internal/conventional"
	"github.com/isokolovskii/commitizen/internal/git"
	"github.com/isokolovskii/commitizen/internal/prompts"
	"github.com/isokolovskii/commitizen/internal/retry"
)

type (
	fakeRunner struct {
		commitMsg   string
		staged      []string
		commitCalls int
	}
)

const (
	testRoot = "/repo"

	expectedWizardMessage  = "feat(api): add feature\n\ndetails\n\nIssue: JIRA-1\n"
	expectedFlagsMessage   = "fix(core): patch bug\n"
	expectedExampleMessage = "feat(core): add hello world handler"

	filePermForTests = 0o600

	testScope        = "api"
	testStagedFile   = "main.go"
	testBody         = "details"
	testIssue        = "JIRA-1"
	testSubject      = "add feature"
	testFlagsTitle   = "patch bug"
	testFixType      = "fix"
	testFlagsMessage = "fix: patch bug\n"
	testShortMessage = "fix: patch bug"
)

var _ git.Runner = (*fakeRunner)(nil)

func (*fakeRunner) Root() (string, error) {
	return testRoot, nil
}

func (fake *fakeRunner) StagedFiles() ([]string, error) {
	return fake.staged, nil
}

func (fake *fakeRunner) Commit(message string) error {
	fake.commitCalls++
	fake.commitMsg = message

	return nil
}

// newCommitDeps wires a fake runner, scripted prompts and throwaway buffers.
func newCommitDeps(t *testing.T, runner *fakeRunner, prompter prompts.Prompter, home string) *CommitDeps {
	t.Helper()

	return &CommitDeps{
		Runner:   runner,
		Prompter: prompter,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
		Root:     testRoot,
		Home:     home,
		CacheDir: t.TempDir(),
	}
}

func stdoutOf(deps *CommitDeps) string {
	return deps.Stdout.(*bytes.Buffer).String() //nolint:forcetypeassert // test helper owns the buffer
}

func TestRunCommitWizardCreatesCommit(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{staged: []string{testStagedFile}}
	prompter := &prompts.Scripted{
		Selects:  []int{0},
		Inputs:   []string{testScope, testSubject, testBody, testIssue},
		Confirms: []bool{false, true},
	}
	deps := newCommitDeps(t, runner, prompter, t.TempDir())

	err := runCommit(deps, &CommitOptions{})

	require.NoError(t, err)
	assert.Equal(t, 1, runner.commitCalls)
	assert.Equal(t, expectedWizardMessage, runner.commitMsg)
	assert.Equal(t, expectedWizardMessage, stdoutOf(deps))
}

func TestRunCommitWizardHonorsSkipQuestions(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	czrc := `{"path":"conventional","skipQuestions":["scope","body","breaking","issues","confirm"]}`
	writeTestFile(t, home, ".czrc", czrc)

	runner := &fakeRunner{staged: []string{testStagedFile}}
	prompter := &prompts.Scripted{
		Selects: []int{0},
		Inputs:  []string{"subject only"},
	}
	deps := newCommitDeps(t, runner, prompter, home)

	err := runCommit(deps, &CommitOptions{})

	require.NoError(t, err)
	assert.Equal(t, "feat: subject only\n", runner.commitMsg)
}

func TestRunCommitFlagsCommitWithoutPrompting(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	deps := newCommitDeps(t, runner, &prompts.Scripted{}, t.TempDir())
	options := &CommitOptions{TypeSet: true}
	options.Commit = stdconventional.Commit{Type: testFixType, Scope: "core", Title: testFlagsTitle}

	err := runCommit(deps, options)

	require.NoError(t, err)
	assert.Equal(t, 1, runner.commitCalls)
	assert.Equal(t, expectedFlagsMessage, runner.commitMsg)
}

func TestRunCommitDryRunPrintsInsteadOfCommitting(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	deps := newCommitDeps(t, runner, &prompts.Scripted{}, t.TempDir())
	options := &CommitOptions{TypeSet: true, DryRun: true}
	options.Commit = stdconventional.Commit{Type: testFixType, Title: testFlagsTitle}

	err := runCommit(deps, options)

	require.NoError(t, err)
	assert.Zero(t, runner.commitCalls)
	assert.Equal(t, testFlagsMessage, stdoutOf(deps))
}

func TestRunCommitHookWritesFileInsteadOfCommitting(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	deps := newCommitDeps(t, runner, &prompts.Scripted{}, t.TempDir())
	hookFile := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	options := &CommitOptions{TypeSet: true, HookFile: hookFile}
	options.Commit = stdconventional.Commit{Type: testFixType, Title: testFlagsTitle}

	err := runCommit(deps, options)

	require.NoError(t, err)
	assert.Zero(t, runner.commitCalls)

	content, err := os.ReadFile(hookFile)
	require.NoError(t, err)
	assert.Equal(t, testFlagsMessage, string(content))
}

func TestRunCommitRetryReplaysCachedCommit(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	deps := newCommitDeps(t, runner, &prompts.Scripted{}, t.TempDir())
	cached := &stdconventional.Commit{Type: "feat", Scope: "api", Title: testSubject, Body: testBody, Issue: testIssue}
	err := retry.NewStore(deps.CacheDir).Save(testRoot, cached)
	require.NoError(t, err)

	err = runCommit(deps, &CommitOptions{Retry: true})

	require.NoError(t, err)
	assert.Equal(t, 1, runner.commitCalls)
	assert.Equal(t, expectedWizardMessage, runner.commitMsg)
}

func TestRunCommitRetryWithoutCacheExits(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	deps := newCommitDeps(t, runner, &prompts.Scripted{}, t.TempDir())

	err := runCommit(deps, &CommitOptions{Retry: true})

	require.Error(t, err)
	assert.Zero(t, runner.commitCalls)
	assert.ErrorIs(t, err, retry.ErrNoCachedCommit)
}

func TestRunCommitNothingStagedExitsBeforePrompting(t *testing.T) {
	t.Parallel()

	runner := &fakeRunner{}
	deps := newCommitDeps(t, runner, &prompts.Scripted{}, t.TempDir())

	err := runCommit(deps, &CommitOptions{})

	require.Error(t, err)
	assert.Zero(t, runner.commitCalls)
	assert.ErrorIs(t, err, ErrNothingStaged)
}

// writeTestFile creates a file with the given content inside dir.
func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()

	err := os.WriteFile(filepath.Join(dir, name), []byte(content), filePermForTests)
	require.NoError(t, err)
}
