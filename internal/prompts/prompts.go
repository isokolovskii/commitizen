// Package prompts abstracts terminal interaction behind a small interface so
// that adapters stay testable without a TTY.
package prompts

import (
	"errors"
	"fmt"
)

type (
	// Default conveys the preselected answer of a confirm prompt without
	// using a bare bool parameter.
	Default int

	// Option describes one selectable entry of a select prompt.
	Option struct {
		// Value is the machine readable result of the selection.
		Value string
		// Label is the human readable text shown to the user.
		Label string
	}

	// Scripted is a Prompter returning canned answers in order, for tests.
	Scripted struct {
		// Selects holds answers consumed by Select calls, in order.
		Selects []int
		// Inputs holds answers consumed by Input and MultiLine calls, in order.
		Inputs []string
		// Confirms holds answers consumed by Confirm calls, in order.
		Confirms []bool

		selectIdx int
		inputIdx  int
		confirmIx int
	}
)

const (
	// DefaultNo preselects the negative answer.
	DefaultNo Default = iota
	// DefaultYes preselects the affirmative answer.
	DefaultYes
)

// ErrScriptExhausted is returned when a Scripted prompter is asked for
// more answers than the script provides.
var ErrScriptExhausted = errors.New("scripted prompter ran out of answers")

// Select returns the next scripted select answer and errors when answers run
// out.
func (scripted *Scripted) Select(_ string, _ []Option, _ int) (int, error) {
	if scripted.selectIdx >= len(scripted.Selects) {
		return 0, ErrScriptExhausted
	}

	answer := scripted.Selects[scripted.selectIdx]
	scripted.selectIdx++

	return answer, nil
}

// Input returns the next scripted text answer and errors when answers run
// out. Validation is still applied, mirroring real prompt behavior.
func (scripted *Scripted) Input(_, _ string, validate func(string) error) (string, error) {
	if scripted.inputIdx >= len(scripted.Inputs) {
		return "", ErrScriptExhausted
	}

	answer := scripted.Inputs[scripted.inputIdx]
	scripted.inputIdx++

	if validate != nil {
		err := validate(answer)
		if err != nil {
			return "", fmt.Errorf("validating answer: %w", err)
		}
	}

	return answer, nil
}

// MultiLine returns the next scripted text answer and errors when answers run
// out.
func (scripted *Scripted) MultiLine(_, _ string) (string, error) {
	answer, err := scripted.Input("", "", nil)
	if err != nil {
		return "", fmt.Errorf("collecting multi line input: %w", err)
	}

	return answer, nil
}

// Confirm returns the next scripted bool answer and errors when answers run
// out.
func (scripted *Scripted) Confirm(_ string, _ Default) (bool, error) {
	if scripted.confirmIx >= len(scripted.Confirms) {
		return false, ErrScriptExhausted
	}

	answer := scripted.Confirms[scripted.confirmIx]
	scripted.confirmIx++

	return answer, nil
}
