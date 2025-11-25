package cmd

import (
	"github.com/urfave/cli/v3"
)

var commands = []*cli.Command{
	version(),
	commit(),
}
