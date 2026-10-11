package gittest

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	stagedFileName = "hello.txt"
)

func TestRepoHelpersRoundTrip(t *testing.T) {
	t.Parallel()

	dir := Repo(t)
	WriteFile(t, dir, stagedFileName, "hi")
	Stage(t, dir, stagedFileName)

	cmd := exec.CommandContext(
		t.Context(),
		gitBinary,
		"diff",
		"--cached",
		"--name-only",
	)
	cmd.Dir = dir
	cmd.Env = env(t)

	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, stagedFileName, strings.TrimSpace(string(out)))
}
