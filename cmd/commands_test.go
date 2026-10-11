package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ver "github.com/isokolovskii/commitizen/internal/version"
)

func TestRunListWritesAdapterLine(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}

	err := runList(out)

	require.NoError(t, err)
	assert.Equal(t, "conventional - Conventional Commits wizard (feat, fix, docs, and friends)\n", out.String())
}

func TestRunExamplePrintsDefaultAdapterExample(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}

	err := runExample(out)

	require.NoError(t, err)
	assert.Truef(t, strings.HasPrefix(out.String(), expectedExampleMessage),
		"example should start with the feat example, got %q", out.String())
}

func TestPrintVersionWritesVersion(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}

	err := printVersion(out)

	require.NoError(t, err)
	assert.Equal(t, ver.Version()+"\n", out.String())
}
