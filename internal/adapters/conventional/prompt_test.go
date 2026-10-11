package conventional

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isokolovskii/commitizen/internal/config"
	"github.com/isokolovskii/commitizen/internal/conventional"
	"github.com/isokolovskii/commitizen/internal/prompts"
)

const (
	wizardScope      = "api"
	wizardSubject    = "add feature"
	wizardBody       = "details"
	wizardIssue      = "JIRA-1"
	wizardBreaking   = "drops the CLI flag"
	wizardWideSubj   = "this subject is far too long for the configured width"
	wizardBlankSubj  = "   "
	fixTypeValue     = "fix"
	customTypeValue  = "sparkle"
	customTypeDesc   = "Add glitter"
	testMaxWidthTiny = 5
)

func newWizard(t *testing.T) (*Adapter, *prompts.Scripted) {
	t.Helper()

	scripted := &prompts.Scripted{}

	return New(config.Default()), scripted
}

func allQuestions() []string {
	return []string{
		config.QuestionScope,
		config.QuestionBody,
		config.QuestionBreaking,
		config.QuestionIssues,
		config.QuestionConfirm,
	}
}

func TestPromptBuildsCommitFromAnswers(t *testing.T) {
	t.Parallel()

	adapter, scripted := newWizard(t)
	scripted.Selects = []int{0}
	scripted.Inputs = []string{wizardScope, wizardSubject, wizardBody, wizardIssue}
	scripted.Confirms = []bool{false, true}

	commit, err := adapter.Prompt(scripted)

	require.NoError(t, err)
	assert.Equal(t, featTypeValue, commit.Type)
	assert.Equal(t, wizardScope, commit.Scope)
	assert.Equal(t, wizardSubject, commit.Title)
	assert.Equal(t, wizardBody, commit.Body)
	assert.Equal(t, wizardIssue, commit.Issue)
	assert.Empty(t, commit.BreakingChange)
}

func TestPromptRecordsBreakingChange(t *testing.T) {
	t.Parallel()

	adapter, scripted := newWizard(t)
	scripted.Selects = []int{1}
	scripted.Inputs = []string{wizardScope, wizardSubject, wizardBody, wizardBreaking, wizardIssue}
	scripted.Confirms = []bool{true, true}

	commit, err := adapter.Prompt(scripted)

	require.NoError(t, err)
	assert.Equal(t, fixTypeValue, commit.Type)
	assert.Equal(t, wizardBreaking, commit.BreakingChange)
}

func TestPromptHonorsSkipQuestions(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.SkipQuestions = allQuestions()

	scripted := &prompts.Scripted{Selects: []int{0}, Inputs: []string{wizardSubject}}

	commit, err := New(cfg).Prompt(scripted)

	require.NoError(t, err)
	assert.Equal(t, &conventional.Commit{Type: featTypeValue, Title: wizardSubject}, commit)
}

func TestPromptSkipsConfirmWhenDisabled(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	confirmOff := false
	cfg.Confirm = &confirmOff

	scripted := &prompts.Scripted{
		Selects:  []int{0},
		Inputs:   []string{wizardScope, wizardSubject, wizardBody, wizardIssue},
		Confirms: []bool{false},
	}

	commit, err := New(cfg).Prompt(scripted)

	require.NoError(t, err)
	assert.Equal(t, wizardSubject, commit.Title)
	assert.Len(t, scripted.Confirms, 1)
}

func TestPromptCanceledOnDecline(t *testing.T) {
	t.Parallel()

	adapter, scripted := newWizard(t)
	scripted.Selects = []int{0}
	scripted.Inputs = []string{wizardScope, wizardSubject, wizardBody, wizardIssue}
	scripted.Confirms = []bool{false, false}

	_, err := adapter.Prompt(scripted)

	require.ErrorIs(t, err, ErrCommitCanceled)
}

func TestPromptRejectsBlankSubject(t *testing.T) {
	t.Parallel()

	adapter, scripted := newWizard(t)
	scripted.Selects = []int{0}
	scripted.Inputs = []string{wizardScope, wizardBlankSubj}

	_, err := adapter.Prompt(scripted)

	require.ErrorIs(t, err, conventional.ErrRequiredPartNotPreset)
}

func TestPromptRejectsTooWideSubject(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.MaxSubjectWidth = testMaxWidthTiny

	scripted := &prompts.Scripted{Selects: []int{0}, Inputs: []string{wizardScope, wizardWideSubj}}

	_, err := New(cfg).Prompt(scripted)

	require.ErrorIs(t, err, ErrSubjectTooLong)
}

func TestPromptWrapsScriptExhaustion(t *testing.T) {
	t.Parallel()

	adapter, scripted := newWizard(t)

	_, err := adapter.Prompt(scripted)

	require.ErrorIs(t, err, prompts.ErrScriptExhausted)
	assert.ErrorContains(t, err, "running wizard step")
}

func TestDefaultTypeIndex(t *testing.T) {
	t.Parallel()

	options := []prompts.Option{{Value: featTypeValue}, {Value: fixTypeValue}}

	tests := []struct {
		defaultType string
		expect      int
	}{
		{defaultType: "", expect: 0},
		{defaultType: "fix", expect: 1},
		{defaultType: "missing", expect: 0},
	}

	for i := range tests {
		test := tests[i]

		t.Run(test.defaultType, func(t *testing.T) {
			t.Parallel()

			cfg := config.Default()
			cfg.DefaultType = test.defaultType

			assert.Equal(t, test.expect, defaultTypeIndex(cfg, options))
		})
	}
}

func TestTypeOptionsComposeLabels(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.Types = []config.Type{{Value: customTypeValue, Description: customTypeDesc}}

	options := New(cfg).typeOptions()

	require.Len(t, options, 1)
	assert.Equal(t, customTypeValue, options[0].Value)
	assert.Equal(t, customTypeValue+": "+customTypeDesc, options[0].Label)
}
