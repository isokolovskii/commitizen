package cmd

import (
	"github.com/urfave/cli/v3"

	ver "github.com/isokolovskii/commitizen/internal/version"
)

func Lefthook() *cli.Command {
	return &cli.Command{
		Name:                  "commitizen",
		Usage:                 "Conventional Commits prompt generator",
		Version:               ver.Version(),
		Commands:              commands,
		EnableShellCompletion: true,
	}
}
