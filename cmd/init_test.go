package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isokolovskii/commitizen/internal/adapters"
	"github.com/isokolovskii/commitizen/internal/config"
)

const (
	czrcGolden  = "{\n  \"path\": \"conventional\"\n}\n"
	czrcStale   = "{\n  \"path\": \"outdated\"\n}\n"
	unknownName = "sparkle"
)

func newInitDeps(t *testing.T) (*InitDeps, *bytes.Buffer) {
	t.Helper()

	out := &bytes.Buffer{}

	return &InitDeps{Dir: t.TempDir(), Stdout: out}, out
}

func TestRunInitWritesDefaultCzrc(t *testing.T) {
	t.Parallel()

	deps, out := newInitDeps(t)

	err := runInit(deps, config.DefaultAdapter, false)

	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(deps.Dir, czrcFileName))
	require.NoError(t, err)
	assert.JSONEq(t, czrcGolden, string(content))
	assert.Equal(t, "Initialized "+filepath.Join(deps.Dir, czrcFileName)+" with adapter conventional\n", out.String())
}

func TestRunInitFailsWhenCzrcExists(t *testing.T) {
	t.Parallel()

	deps, _ := newInitDeps(t)
	writeTestFile(t, deps.Dir, czrcFileName, czrcGolden)

	err := runInit(deps, config.DefaultAdapter, false)

	require.ErrorIs(t, err, ErrConfigExists)

	content, err := os.ReadFile(filepath.Join(deps.Dir, czrcFileName))
	require.NoError(t, err)
	assert.JSONEq(t, czrcGolden, string(content))
}

func TestRunInitForceOverwritesCzrc(t *testing.T) {
	t.Parallel()

	deps, _ := newInitDeps(t)
	writeTestFile(t, deps.Dir, czrcFileName, czrcStale)

	err := runInit(deps, config.DefaultAdapter, true)

	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(deps.Dir, czrcFileName))
	require.NoError(t, err)
	assert.JSONEq(t, czrcGolden, string(content))
}

func TestRunInitRejectsUnknownAdapter(t *testing.T) {
	t.Parallel()

	deps, _ := newInitDeps(t)

	err := runInit(deps, unknownName, false)

	require.Error(t, err)
	require.ErrorIs(t, err, adapters.ErrUnknownAdapter)
	assert.Contains(t, err.Error(), "available")
}
