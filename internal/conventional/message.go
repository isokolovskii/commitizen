package conventional

import (
	"errors"
	"fmt"
)

const (
	newLine = "\n"
)

// ErrRequiredPartNotPreset - error used for missing required conventional commit fields.
var ErrRequiredPartNotPreset = errors.New("required part not present")

// BuildCommitMessage - builds conventional commit message.
func BuildCommitMessage(commit *Commit) (string, error) {
	if commit.Type == "" {
		return "", fmt.Errorf("%w: type", ErrRequiredPartNotPreset)
	}

	if commit.Title == "" {
		return "", fmt.Errorf("%w: title", ErrRequiredPartNotPreset)
	}

	header := buildHeader(commit)
	footer := buildFooter(commit)

	return header + newLine + footer, nil
}

func buildHeader(commit *Commit) string {
	header := commit.Type

	if commit.Scope != "" {
		header = fmt.Sprintf("%s(%s)", header, commit.Scope)
	}

	if commit.BreakingChange != "" {
		header += "!"
	}

	header += ": " + commit.Title

	return header
}

func buildFooter(commit *Commit) string {
	footer := ""

	if commit.BreakingChange != "" {
		footer += newLine + "BREAKING CHANGE: " + commit.BreakingChange
	}

	if commit.Issue != "" {
		footer += newLine + "Issue: " + commit.Issue
	}

	if commit.Body != "" {
		if footer != "" {
			footer = newLine + commit.Body + newLine + footer + newLine
		} else {
			footer = newLine + commit.Body + newLine
		}
	} else if footer != "" {
		footer += newLine
	}

	return footer
}
