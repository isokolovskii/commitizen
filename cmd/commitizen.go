package cmd

import (
	"github.com/urfave/cli/v3"

	ver "github.com/isokolovskii/commitizen/internal/version"
)

// New returns the root commitizen command with commit as the default action.
func New() *cli.Command {
	return &cli.Command{
		Name:                  appName,
		Usage:                 "Conventional Commits prompt generator",
		Version:               ver.Version(),
		DefaultCommand:        commitCommand,
		Commands:              commands,
		EnableShellCompletion: true,
	}
}
