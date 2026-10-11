package prompts

import (
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/huh"
)

type (
	// Interactive is a Prompter backed by charmbracelet/huh forms.
	Interactive struct{}
)

var (
	// ErrUnknownSelection is returned when huh reports a value outside the
	// offered option list.
	ErrUnknownSelection = errors.New("unknown selection")

	_ Prompter = Interactive{}
)

// Select renders a huh select field bound to the option values.
func (Interactive) Select(message string, options []Option, defaultIndex int) (int, error) {
	values := make([]string, 0, len(options))
	for _, option := range options {
		values = append(values, option.Value)
	}

	selected := values[defaultIndex]

	err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title(message).
			Options(huhOptions(options)...).
			Value(&selected),
	)).Run()
	if err != nil {
		return 0, fmt.Errorf("selecting option: %w", err)
	}

	index := slices.Index(values, selected)
	if index < 0 {
		return 0, fmt.Errorf("%w: %q", ErrUnknownSelection, selected)
	}

	return index, nil
}

// Input renders a huh input field, optionally validated.
func (Interactive) Input(message, defaultValue string, validate func(string) error) (string, error) {
	answer := defaultValue

	field := huh.NewInput().Title(message).Value(&answer)
	if validate != nil {
		field = field.Validate(validate)
	}

	err := huh.NewForm(huh.NewGroup(field)).Run()
	if err != nil {
		return "", fmt.Errorf("collecting input: %w", err)
	}

	return answer, nil
}

// MultiLine renders a huh text editor field.
func (Interactive) MultiLine(message, defaultValue string) (string, error) {
	answer := defaultValue

	err := huh.NewForm(huh.NewGroup(
		huh.NewText().Title(message).Value(&answer),
	)).Run()
	if err != nil {
		return "", fmt.Errorf("collecting multiline description: %w", err)
	}

	return answer, nil
}

// Confirm renders a huh confirm field with the given preselected answer.
func (Interactive) Confirm(message string, preselected Default) (bool, error) {
	confirmed := preselected == DefaultYes

	err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title(message).Value(&confirmed),
	)).Run()
	if err != nil {
		return false, fmt.Errorf("confirming: %w", err)
	}

	return confirmed, nil
}

// huhOptions converts prompt options into huh select options.
func huhOptions(options []Option) []huh.Option[string] {
	result := make([]huh.Option[string], 0, len(options))
	for _, option := range options {
		result = append(result, huh.NewOption(option.Label, option.Value))
	}

	return result
}
