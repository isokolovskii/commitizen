package cmd

import (
	"github.com/urfave/cli/v3"
)

// commandSet builds fresh command instances so every root command owns its
// own flag state.
func commandSet() []*cli.Command {
	return []*cli.Command{
		commit(),
		initCommand(),
		listCommand(),
		exampleCommand(),
		versionCommand(),
	}
}
