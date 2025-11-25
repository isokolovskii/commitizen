package conventional

import (
	"errors"
	"fmt"
)

var ErrRequiredPartNotPreset = errors.New("required part not present")

func BuildCommit(commit *Commit) (string, error) {
	result := commit.Type

	if commit.Type == "" {
		return "", fmt.Errorf("%w: type", ErrRequiredPartNotPreset)
	}

	if commit.Title == "" {
		return "", fmt.Errorf("%w: title", ErrRequiredPartNotPreset)
	}

	if commit.Scope != "" {
		result = fmt.Sprintf("%s(%s)", result, commit.Scope)
	}

	result = fmt.Sprintf("%s: %s", result, commit.Title)

	if commit.Body != "" {
		result = fmt.Sprintf("%s\n\n%s", result, commit.Body)
	}

	if commit.Footer != "" {
		result = fmt.Sprintf("%s\n\n%s\n", result, commit.Footer)
	}

	return result, nil
}
