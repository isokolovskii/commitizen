package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	ver "github.com/isokolovskii/commitizen/internal/version"
)

func versionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "print the commitizen version",
		Action: func(_ context.Context, cmd *cli.Command) error {
			err := printVersion(cmd.Root().Writer)
			if err != nil {
				return fmt.Errorf("running version command: %w", err)
			}

			return nil
		},
	}
}

// printVersion writes the current version to out.
func printVersion(out io.Writer) error {
	_, err := fmt.Fprintln(out, ver.Version())
	if err != nil {
		return fmt.Errorf("printing version: %w", err)
	}

	return nil
}
