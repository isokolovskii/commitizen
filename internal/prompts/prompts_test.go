package prompts_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isokolovskii/commitizen/internal/prompts"
)

const (
	msgType    = "type"
	msgSubject = "subject"
	msgBody    = "body"
	msgBodyful = "body line"
	msgProceed = "proceed"
	answerOK   = "ok"
)

var errEmptyAnswer = errors.New("answer must not be empty")

func TestScriptedSelectConsumesAnswersInOrder(t *testing.T) {
	t.Parallel()

	prompter := &prompts.Scripted{Selects: []int{2, 0}}

	options := []prompts.Option{{Value: "a"}, {Value: "b"}, {Value: "c"}}

	first, err := prompter.Select(msgType, options, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, first)

	second, err := prompter.Select(msgType, options, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, second)
}

func TestScriptedSelectExhausted(t *testing.T) {
	t.Parallel()

	prompter := &prompts.Scripted{}

	_, err := prompter.Select(msgType, nil, 0)

	require.ErrorIs(t, err, prompts.ErrScriptExhausted)
}

func TestScriptedInputAppliesValidation(t *testing.T) {
	t.Parallel()

	prompter := &prompts.Scripted{Inputs: []string{answerOK, ""}}
	validate := func(answer string) error {
		if answer == "" {
			return errEmptyAnswer
		}

		return nil
	}

	answer, err := prompter.Input(msgSubject, "", validate)
	require.NoError(t, err)
	assert.Equal(t, answerOK, answer)

	_, err = prompter.Input(msgSubject, "", validate)

	require.ErrorIs(t, err, errEmptyAnswer)
}

func TestScriptedInputExhausted(t *testing.T) {
	t.Parallel()

	prompter := &prompts.Scripted{}

	_, err := prompter.Input(msgSubject, "", nil)

	require.ErrorIs(t, err, prompts.ErrScriptExhausted)
}

func TestScriptedMultiLineSharesInputQueue(t *testing.T) {
	t.Parallel()

	prompter := &prompts.Scripted{Inputs: []string{msgBodyful}}

	answer, err := prompter.MultiLine(msgBody, "")
	require.NoError(t, err)
	assert.Equal(t, msgBodyful, answer)

	_, err = prompter.MultiLine(msgBody, "")

	require.ErrorIs(t, err, prompts.ErrScriptExhausted)
}

func TestScriptedConfirmConsumesAnswersInOrder(t *testing.T) {
	t.Parallel()

	prompter := &prompts.Scripted{Confirms: []bool{false, true}}

	first, err := prompter.Confirm(msgProceed, prompts.DefaultYes)
	require.NoError(t, err)
	assert.False(t, first)

	second, err := prompter.Confirm(msgProceed, prompts.DefaultYes)
	require.NoError(t, err)
	assert.True(t, second)
}

func TestScriptedConfirmExhausted(t *testing.T) {
	t.Parallel()

	prompter := &prompts.Scripted{}

	_, err := prompter.Confirm(msgProceed, prompts.DefaultNo)

	require.ErrorIs(t, err, prompts.ErrScriptExhausted)
}
