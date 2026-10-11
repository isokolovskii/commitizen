package retry_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isokolovskii/commitizen/internal/conventional"
	"github.com/isokolovskii/commitizen/internal/retry"
)

const (
	repoRoot   = "/repos/commitizen"
	otherRoot  = "/repos/first"
	firstTitle = "first"
	fixType    = "fix"
)

func commitFixture() *conventional.Commit {
	return &conventional.Commit{
		Type:           "feat",
		Scope:          "wizard",
		Title:          "cache answers",
		Body:           "Store answers between runs.",
		BreakingChange: "cache file moves",
		Issue:          "TEST-7",
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	t.Parallel()

	store := retry.NewStore(t.TempDir())
	want := commitFixture()

	require.NoError(t, store.Save(repoRoot, want))

	got, err := store.Load(repoRoot)

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestLoadWithoutCacheFile(t *testing.T) {
	t.Parallel()

	store := retry.NewStore(t.TempDir())

	_, err := store.Load(repoRoot)

	require.ErrorIs(t, err, retry.ErrNoCachedCommit)
}

func TestLoadUnknownRepository(t *testing.T) {
	t.Parallel()

	store := retry.NewStore(t.TempDir())
	require.NoError(t, store.Save(repoRoot, commitFixture()))

	_, err := store.Load("/repos/other")

	require.ErrorIs(t, err, retry.ErrNoCachedCommit)
}

func TestSaveOverwritesEntry(t *testing.T) {
	t.Parallel()

	store := retry.NewStore(t.TempDir())
	require.NoError(t, store.Save(repoRoot, &conventional.Commit{Type: fixType, Title: firstTitle}))

	want := commitFixture()
	require.NoError(t, store.Save(repoRoot, want))

	got, err := store.Load(repoRoot)

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestSaveKeepsOtherRepositories(t *testing.T) {
	t.Parallel()

	store := retry.NewStore(t.TempDir())
	first := &conventional.Commit{Type: fixType, Title: firstTitle}
	require.NoError(t, store.Save(otherRoot, first))

	require.NoError(t, store.Save(repoRoot, commitFixture()))

	got, err := store.Load(otherRoot)

	require.NoError(t, err)
	assert.Equal(t, first, got)
}

func TestLoadRejectsCorruptCacheFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	store := retry.NewStore(dir)
	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "retry.json"), []byte("{not json"), 0o600))

	_, err := store.Load(repoRoot)

	require.Error(t, err)
	assert.ErrorContains(t, err, "decoding")
}

func TestSaveCreatesNestedCacheDirectory(t *testing.T) {
	t.Parallel()

	store := retry.NewStore(filepath.Join(t.TempDir(), "nested", "deeper"))

	require.NoError(t, store.Save(repoRoot, commitFixture()))
}
