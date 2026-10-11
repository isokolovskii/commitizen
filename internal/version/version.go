package version

import (
	"errors"

	"golang.org/x/mod/semver"
)

const (
	version = "1.0.0"
)

var (
	// Is set via -X github.com/isokolovskii/commitizen/internal/version.commit={commit}.
	commit string

	ErrInvalidVersion   = errors.New("invalid version format")
	ErrUncoveredVersion = errors.New("version is lower than required")
)

// Version reports the human readable version, optionally suffixed with the git commit injected at build time.
func Version() string {
	if commit == "" {
		return version
	}

	return version + " " + commit
}

func Check(wanted, given string) error {
	if wanted[0] != 'v' {
		wanted = "v" + wanted
	}
	if given[0] != 'v' {
		given = "v" + given
	}

	if !semver.IsValid(wanted) {
		return ErrInvalidVersion
	}
	if !semver.IsValid(given) {
		return ErrInvalidVersion
	}
	if cmp := semver.Compare(wanted, given); cmp > 0 {
		return ErrUncoveredVersion
	}

	return nil
}
