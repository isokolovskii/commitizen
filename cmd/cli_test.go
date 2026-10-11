package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/isokolovskii/commitizen/internal/adapters"
	conventionaladapter "github.com/isokolovskii/commitizen/internal/adapters/conventional"
	"github.com/isokolovskii/commitizen/internal/git"
	"github.com/isokolovskii/commitizen/internal/git/gittest"
	"github.com/isokolovskii/commitizen/internal/prompts"
	"github.com/isokolovskii/commitizen/internal/retry"
	ver "github.com/isokolovskii/commitizen/internal/version"
)

type (
	// ErrWriter always fails, standing in for a broken stdout.
	errWriter struct{}
)

const (
	cliExitCode          = 1
	initCommandName      = "init"
	blockerName          = "blocker"
	msgUnknownAdapter    = "unknown adapter"
	msgNotInsideRepo     = "not inside a git repository"
	msgNoCachedToRetry   = "no cached commit to retry"
	msgNothingStagedHint = "nothing staged"
)

var errWriteFailed = errors.New("write failed")

func (errWriter) Write([]byte) (int, error) {
	return 0, errWriteFailed
}

// wrappedUnknown builds the wrapped unknown-adapter failure the wizard produces.
func wrappedUnknown() error {
	return fmt.Errorf("resolving adapter %q: %w", unknownName, adapters.ErrUnknownAdapter)
}

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()

	root := New()
	out := &bytes.Buffer{}
	root.Writer = out

	err := root.Run(t.Context(), append([]string{appName}, args...))

	return out.String(), err
}

// captureExits swaps the process exit hook so CLI failures become assertions.
func captureExits(t *testing.T) *[]int {
	t.Helper()

	codes := &[]int{}
	original := cli.OsExiter
	cli.OsExiter = func(code int) {
		*codes = append(*codes, code)
	}
	t.Cleanup(func() { cli.OsExiter = original })

	return codes
}

func exitCodeOf(t *testing.T, err error) int {
	t.Helper()

	exit, ok := err.(cli.ExitCoder) //nolint:errorlint // asserting the terminal CLI error type
	require.True(t, ok)

	return exit.ExitCode()
}

func TestCLIListsAdapters(t *testing.T) {
	t.Parallel()

	out, err := runCLI(t, "ls")

	require.NoError(t, err)
	assert.Contains(t, out, testAdapterLine)
}

func TestCLIPrintsExample(t *testing.T) {
	t.Parallel()

	out, err := runCLI(t, "example")

	require.NoError(t, err)
	assert.Contains(t, out, expectedExampleMessage)
}

func TestCLIPrintsVersion(t *testing.T) {
	t.Parallel()

	out, err := runCLI(t, "version")

	require.NoError(t, err)
	assert.Equal(t, ver.Version()+"\n", out)
}

// The t.Chdir based tests stay sequential: Chdir is incompatible with parallel
// subtests, so the paralleltest exception is intentional on each of them.

func TestCLIDryRunCommitPrintsWithoutRepository(t *testing.T) { //nolint:paralleltest // t.Chdir cannot run in parallel
	t.Chdir(t.TempDir())

	out, err := runCLI(t, commitCommand, testTypeFlag, testFixType, "--title", testDrySubject, "--dry-run")

	require.NoError(t, err)
	assert.Contains(t, out, testDryMessage)
}

func TestCLIInitWritesCzrc(t *testing.T) { //nolint:paralleltest // t.Chdir cannot run in parallel
	dir := t.TempDir()
	t.Chdir(dir)

	out, err := runCLI(t, initCommandName, "conventional")

	require.NoError(t, err)
	assert.Contains(t, out, testInitializedOutput)

	content, err := os.ReadFile(filepath.Join(dir, czrcFileName))
	require.NoError(t, err)
	assert.JSONEq(t, `{"path":"conventional"}`, string(content))
}

func TestCLIInitUnknownAdapterExits(t *testing.T) { //nolint:paralleltest // t.Chdir cannot run in parallel
	t.Chdir(t.TempDir())
	codes := captureExits(t)

	_, err := runCLI(t, initCommandName, unknownName)

	require.Error(t, err)
	assert.Contains(t, err.Error(), msgUnknownAdapter)
	assert.Equal(t, []int{cliExitCode}, *codes)
}

func TestCLICommitOutsideRepositoryExits(t *testing.T) { //nolint:paralleltest // t.Chdir cannot run in parallel
	t.Chdir(t.TempDir())
	codes := captureExits(t)

	_, err := runCLI(t, commitCommand)

	require.Error(t, err)
	assert.Contains(t, err.Error(), msgNotInsideRepo)
	assert.Equal(t, []int{cliExitCode}, *codes)
}

func TestCLIRetryWithoutCacheExits(t *testing.T) { //nolint:paralleltest // t.Chdir cannot run in parallel
	t.Chdir(gittest.Repo(t))
	codes := captureExits(t)

	_, err := runCLI(t, commitCommand, "--retry")

	require.Error(t, err)
	assert.Contains(t, err.Error(), msgNoCachedToRetry)
	assert.Equal(t, []int{cliExitCode}, *codes)
}

func TestMapCommitErrorTranslatesKnownFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err       error
		expectMsg string
	}{
		{err: git.ErrNotRepository, expectMsg: msgNotInsideRepo},
		{err: ErrNothingStaged, expectMsg: msgNothingStagedHint},
		{err: retry.ErrNoCachedCommit, expectMsg: msgNoCachedToRetry},
		{err: conventionaladapter.ErrCommitCanceled, expectMsg: "commit canceled"},
		{err: wrappedUnknown(), expectMsg: msgUnknownAdapter},
		{err: filepath.ErrBadPattern, expectMsg: filepath.ErrBadPattern.Error()},
	}

	for i := range tests {
		test := tests[i]

		t.Run(test.expectMsg, func(t *testing.T) {
			t.Parallel()

			err := mapCommitError(test.err)

			require.Error(t, err)
			assert.Equal(t, cliExitCode, exitCodeOf(t, err))
			assert.Contains(t, err.Error(), test.expectMsg)
		})
	}
}

func TestMapInitErrorTranslatesKnownFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err        error
		expectMsg  string
		expectExit bool
	}{
		{err: ErrConfigExists, expectMsg: "configuration file already exists", expectExit: true},
		{err: wrappedUnknown(), expectMsg: msgUnknownAdapter, expectExit: true},
		{err: filepath.ErrBadPattern, expectMsg: filepath.ErrBadPattern.Error()},
	}

	for i := range tests {
		test := tests[i]

		t.Run(test.expectMsg, func(t *testing.T) {
			t.Parallel()

			err := mapInitError(test.err)

			require.Error(t, err)
			assert.Contains(t, err.Error(), test.expectMsg)
			if !test.expectExit {
				assert.NotErrorIs(t, err, ErrConfigExists)

				return
			}

			assert.Equal(t, cliExitCode, exitCodeOf(t, err))
		})
	}
}

func TestUserCacheDirAppendsNamespace(t *testing.T) {
	t.Parallel()

	dir := userCacheDir()

	assert.Truef(t, filepath.IsAbs(dir), "cache dir %q should be absolute", dir)
	assert.Equal(t, appName, filepath.Base(dir))
}

func TestPrintCommitMessageWriterError(t *testing.T) {
	t.Parallel()

	err := printCommitMessage(errWriter{}, "feat: x\n")

	require.ErrorIs(t, err, errWriteFailed)
	assert.ErrorContains(t, err, "printing commit message")
}

func TestSaveRetryWarnsWhenCacheUnwritable(t *testing.T) {
	t.Parallel()

	blocker := filepath.Join(t.TempDir(), blockerName)
	writeTestFile(t, filepath.Dir(blocker), blockerName, "not a directory")

	runner := &fakeRunner{staged: []string{testStagedFile}}
	scripted := &prompts.Scripted{
		Selects:  []int{0},
		Inputs:   []string{testScope, testSubject, testBody, testIssue},
		Confirms: []bool{false, true},
	}
	deps := newCommitDeps(t, runner, scripted, t.TempDir())
	deps.CacheDir = filepath.Join(blocker, "cache")

	err := runCommit(deps, &CommitOptions{})

	require.NoError(t, err)

	stderr, ok := deps.Stderr.(*bytes.Buffer)
	require.True(t, ok)
	assert.Contains(t, stderr.String(), "warning: could not save retry cache")
	assert.Contains(t, runner.commitMsg, testSubject)
}

func TestWizardSurfacesAdapterLookupFailure(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	writeTestFile(t, home, czrcFileName, `{"path":"sparkle"}`)

	runner := &fakeRunner{staged: []string{testStagedFile}}
	deps := newCommitDeps(t, runner, &prompts.Scripted{}, home)

	_, err := wizardMessage(deps)

	require.Error(t, err)
	assert.ErrorIs(t, err, adapters.ErrUnknownAdapter)
}
