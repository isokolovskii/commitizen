package prompts_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isokolovskii/commitizen/internal/prompts"
)

// The huh backend needs a TTY, which test environments never provide. Each
// method therefore exercises its error path, asserting that failures surface
// with the operation context attached.

const (
	ttyMarker  = "could not open a new TTY"
	huhFeat    = "feat"
	huhOptions = "options"
)

func TestInteractiveSelectReportsTTYError(t *testing.T) {
	t.Parallel()

	_, err := prompts.Interactive{}.Select(msgType, []prompts.Option{{Value: huhFeat, Label: huhFeat}}, 0)

	require.Error(t, err)
	require.ErrorContains(t, err, "selecting option")
	assert.Containsf(t, err.Error(), ttyMarker, "expected TTY failure, got %q", err.Error())
}

func TestInteractiveInputReportsTTYError(t *testing.T) {
	t.Parallel()

	_, err := prompts.Interactive{}.Input(msgSubject, "", nil)

	require.Error(t, err)
	assert.ErrorContains(t, err, "collecting input")
}

func TestInteractiveMultiLineReportsTTYError(t *testing.T) {
	t.Parallel()

	_, err := prompts.Interactive{}.MultiLine(msgBody, "")

	require.Error(t, err)
	assert.ErrorContains(t, err, "collecting multiline description")
}

func TestInteractiveConfirmReportsTTYError(t *testing.T) {
	t.Parallel()

	_, err := prompts.Interactive{}.Confirm(msgProceed, prompts.DefaultYes)

	require.Error(t, err)
	assert.ErrorContains(t, err, "confirming")
}
