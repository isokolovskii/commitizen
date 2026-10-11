package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isokolovskii/commitizen/internal/version"
)

const (
	e2eModule     = "github.com/isokolovskii/commitizen"
	e2eGoBinary   = "go"
	e2eGitBinary  = "git"
	e2eBinaryName = "commitizen"
	e2eCommitName = "commit"

	e2eConfigName    = "config"
	e2eVersionName   = "version"
	e2eLsName        = "ls"
	e2eExampleName   = "example"
	e2eTitleFlag     = "--title"
	e2eDryRunFlag    = "--dry-run"
	e2eRetryFlag     = "--retry"
	e2eAdapter       = "conventional"
	e2eRepoDir       = "repo"
	e2eStagedFile    = "feature.txt"
	e2eStagedContent = "data"
	e2eAddName       = "add"
	e2eFeatMessage   = "feat: e2e subject"
	e2eSubject       = "e2e subject"
	e2eRetryTitle    = "replayed subject"
	e2eIdentity      = "Test User"
	e2eEmail         = "test@example.com"
	e2eLogName       = "log"
	e2eLogCount      = "-1"
	e2ePrettyBody    = "--pretty=%B"
	e2eDirPerm       = 0o750
)

// buildBinary compiles the module into a fresh executable.
func buildBinary(t *testing.T) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), e2eBinaryName)
	build := exec.CommandContext(t.Context(), e2eGoBinary, "build", "-o", bin, e2eModule)
	output, err := build.CombinedOutput()
	require.NoErrorf(t, err, "go build failed: %s", output)

	return bin
}

// childEnv pins HOME and the git identity so runs are reproducible.
func childEnv(work string) []string {
	return append(os.Environ(),
		"HOME="+work,
		"XDG_CONFIG_HOME="+filepath.Join(work, ".config"),
		testGitNoSystem,
		"GIT_AUTHOR_NAME="+e2eIdentity,
		"GIT_AUTHOR_EMAIL="+e2eEmail,
		"GIT_COMMITTER_NAME="+e2eIdentity,
		"GIT_COMMITTER_EMAIL="+e2eEmail,
	)
}

// newE2E builds the binary plus an isolated workspace with pinned env.
func newE2E(t *testing.T) (string, string, []string) {
	t.Helper()

	work := t.TempDir()

	return buildBinary(t), work, childEnv(work)
}

// runBinary executes the built CLI in dir and returns stdout with stderr.
func runBinary(t *testing.T, bin, dir string, env []string, args ...string) (string, error) {
	t.Helper()

	runner := exec.CommandContext(t.Context(), bin, args...)
	runner.Dir = dir
	runner.Env = env

	output := &strings.Builder{}
	runner.Stdout = output
	runner.Stderr = output

	err := runner.Run()

	return output.String(), err
}

// gitRun runs a git command in dir, failing the test on error.
func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()

	runner := exec.CommandContext(t.Context(), e2eGitBinary, args...)
	runner.Dir = dir
	runner.Env = append(os.Environ(), testGitNoSystem)

	output, err := runner.CombinedOutput()
	require.NoErrorf(t, err, "git %v failed: %s", args, output)

	return string(output)
}

// gitLog returns the last commit message without git's trailing newline.
func gitLog(t *testing.T, repo string) string {
	t.Helper()

	return strings.TrimRight(gitRun(t, repo, e2eLogName, e2eLogCount, e2ePrettyBody), "\n")
}

// newRepo creates a real repository with one staged file.
func newRepo(t *testing.T) string {
	t.Helper()

	repo := filepath.Join(t.TempDir(), e2eRepoDir)
	require.NoError(t, os.MkdirAll(repo, e2eDirPerm))

	for _, args := range [][]string{
		{initCommandName},
		{e2eConfigName, "user.name", e2eIdentity},
		{e2eConfigName, "user.email", e2eEmail},
	} {
		gitRun(t, repo, args...)
	}

	require.NoError(t, os.WriteFile(filepath.Join(repo, e2eStagedFile), []byte(e2eStagedContent), filePermForTests))
	gitRun(t, repo, e2eAddName, e2eStagedFile)

	return repo
}

// cacheNamespace mirrors os.UserCacheDir for the pinned work directory.
func cacheNamespace(work string) string {
	if runtime.GOOS == "darwin" {
		return filepath.Join(work, "Library", "Caches", e2eBinaryName)
	}

	return filepath.Join(work, ".cache", e2eBinaryName)
}

func TestE2EVersion(t *testing.T) {
	t.Parallel()

	bin, work, env := newE2E(t)

	out, err := runBinary(t, bin, work, env, e2eVersionName)

	require.NoError(t, err)
	assert.Equal(t, version.Version()+"\n", out)
}

func TestE2EListsAdapters(t *testing.T) {
	t.Parallel()

	bin, work, env := newE2E(t)

	out, err := runBinary(t, bin, work, env, e2eLsName)

	require.NoError(t, err)
	assert.Contains(t, out, testAdapterLine)
}

func TestE2EPrintsExample(t *testing.T) {
	t.Parallel()

	bin, work, env := newE2E(t)

	out, err := runBinary(t, bin, work, env, e2eExampleName)

	require.NoError(t, err)
	assert.Contains(t, out, expectedExampleMessage)
}

func TestE2EDryRunOutsideRepository(t *testing.T) {
	t.Parallel()

	bin, work, env := newE2E(t)

	out, err := runBinary(
		t,
		bin,
		work,
		env,
		e2eCommitName,
		testTypeFlag,
		"fix",
		e2eTitleFlag,
		testDrySubject,
		e2eDryRunFlag,
	)

	require.NoError(t, err)
	assert.Contains(t, out, testDryMessage)
}

func TestE2EInitWritesCzrc(t *testing.T) {
	t.Parallel()

	bin, work, env := newE2E(t)

	out, err := runBinary(t, bin, work, env, initCommandName, e2eAdapter)

	require.NoError(t, err)
	assert.Contains(t, out, testInitializedOutput)

	content, err := os.ReadFile(filepath.Join(work, czrcFileName))
	require.NoError(t, err)
	assert.JSONEq(t, czrcGolden, string(content))
}

func TestE2EFlagsCommit(t *testing.T) {
	t.Parallel()

	bin, _, env := newE2E(t)
	repo := newRepo(t)

	out, err := runBinary(t, bin, repo, env, e2eCommitName, testTypeFlag, testFeatType, e2eTitleFlag, e2eSubject)

	require.NoError(t, err)
	assert.Contains(t, out, e2eFeatMessage)
	assert.Equal(t, e2eFeatMessage, gitLog(t, repo))
}

func TestE2ERetryReplaysCachedCommit(t *testing.T) {
	t.Parallel()

	bin, work, env := newE2E(t)
	repo := newRepo(t)

	root := strings.TrimSpace(gitRun(t, repo, "rev-parse", "--show-toplevel"))
	cacheDir := cacheNamespace(work)
	require.NoError(t, os.MkdirAll(cacheDir, e2eDirPerm))

	cached := `{"` + root + `":{"type":"feat","title":"` + e2eRetryTitle + `","issue":"E2E-1"}}`
	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "retry.json"), []byte(cached), filePermForTests))

	out, err := runBinary(t, bin, repo, env, e2eCommitName, e2eRetryFlag)

	require.NoError(t, err)
	assert.Contains(t, out, e2eRetryTitle)
	assert.Equal(t, "feat: "+e2eRetryTitle+"\n\nIssue: E2E-1", gitLog(t, repo))
}
