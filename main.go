package main

import (
	"context"
	"fmt"
	"os"

	"github.com/isokolovskii/commitizen/cmd"
)

func main() {
	err := cmd.New().Run(context.Background(), os.Args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
