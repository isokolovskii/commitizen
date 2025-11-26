package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/isokolovskii/commitizen/internal/conventional"
)

var ErrInvalidFlag = errors.New("invalid flag")

func commit() *cli.Command {
	commitData := conventional.Commit{}

	return &cli.Command{
		Name:  "commit",
		Usage: "Create Conventional Commit",
		Flags: flags(&commitData),
		Action: func(_ context.Context, _ *cli.Command) error {
			message, err := conventional.BuildCommitMessage(&commitData)
			if err != nil {
				return fmt.Errorf("error building commit: %w", err)
			}

			_, err = fmt.Println(message)
			if err != nil {
				return fmt.Errorf("error printing built commit message: %w", err)
			}

			return nil
		},
	}
}

func flags(commitData *conventional.Commit) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "type",
			OnlyOnce:    true,
			Destination: &commitData.Type,
			Required:    true,
			Usage:       "Type of change (e.g., feat, fix, docs)",
		},
		&cli.StringFlag{
			Name:        "scope",
			OnlyOnce:    true,
			Destination: &commitData.Scope,
			Required:    false,
			Usage:       "Optional context for the change (e.g., api, cli)",
		},
		&cli.StringFlag{
			Name:        "title",
			OnlyOnce:    true,
			Destination: &commitData.Title,
			Required:    true,
			Usage:       "Short description of changes",
		},
		&cli.StringFlag{
			Name:        "body",
			OnlyOnce:    true,
			Destination: &commitData.Body,
			Required:    false,
			Usage:       "Optional longer description of the change",
		},
		&cli.StringFlag{
			Name:        "breaking",
			OnlyOnce:    true,
			Destination: &commitData.BreakingChange,
			Required:    false,
			Usage:       "Optional description of breaking changes introduced with commit",
		},
		&cli.StringFlag{
			Name:        "issue",
			OnlyOnce:    true,
			Destination: &commitData.Issue,
			Required:    false,
			Usage:       "Optional issue number",
		},
	}
}
