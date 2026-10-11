package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/isokolovskii/commitizen/internal/config"
)

func listCommand() *cli.Command {
	return &cli.Command{
		Name:  "ls",
		Usage: "list the built-in adapters",
		Action: func(_ context.Context, cmd *cli.Command) error {
			err := runList(cmd.Root().Writer)
			if err != nil {
				return fmt.Errorf("running ls command: %w", err)
			}

			return nil
		},
	}
}

// runList writes one line per built-in adapter to out.
func runList(out io.Writer) error {
	registry, err := builtinRegistry(config.Default())
	if err != nil {
		return fmt.Errorf(errBuildRegistry, err)
	}

	for _, name := range registry.Names() {
		adapter, err := registry.Get(name)
		if err != nil {
			return fmt.Errorf("resolving adapter %s: %w", name, err)
		}

		_, err = fmt.Fprintf(out, "%s - %s\n", name, adapter.Description())
		if err != nil {
			return fmt.Errorf("listing adapter %s: %w", name, err)
		}
	}

	return nil
}
