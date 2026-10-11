package cmd

import (
	"github.com/urfave/cli/v3"
)

var commands = []*cli.Command{
	commit(),
	initCommand(),
	listCommand(),
	exampleCommand(),
	versionCommand(),
}
