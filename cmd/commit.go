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
	commit := conventional.Commit{}

	return &cli.Command{
		Name:  "commit",
		Usage: "Create Conventional Commit",
		Flags: flags(&commit),
		Action: func(_ context.Context, _ *cli.Command) error {
			commit, err := conventional.BuildCommitMessage(&commit)
			if err != nil {
				return fmt.Errorf("error building commit: %w", err)
			}

			_, err = fmt.Println(commit)
			if err != nil {
				return fmt.Errorf("error printing built commit message: %w", err)
			}

			return nil
		},
	}
}

func flags(commit *conventional.Commit) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "type",
			OnlyOnce:    true,
			Destination: &commit.Type,
			Required:    true,
			Usage:       "Type of change (e.g., feat, fix, docs)",
		},
		&cli.StringFlag{
			Name:        "scope",
			OnlyOnce:    true,
			Destination: &commit.Scope,
			Required:    false,
			Usage:       "Optional context for the change (e.g., api, cli)",
		},
		&cli.StringFlag{
			Name:        "title",
			OnlyOnce:    true,
			Destination: &commit.Title,
			Required:    true,
			Usage:       "Short description of changes",
		},
		&cli.StringFlag{
			Name:        "body",
			OnlyOnce:    true,
			Destination: &commit.Body,
			Required:    false,
			Usage:       "Optional longer description of the change",
		},
		&cli.StringFlag{
			Name:        "breaking",
			OnlyOnce:    true,
			Destination: &commit.BreakingChange,
			Required:    false,
			Usage:       "Optional description of breaking changes introduced with commit",
		},
		&cli.StringFlag{
			Name:        "issue",
			OnlyOnce:    true,
			Destination: &commit.Issue,
			Required:    false,
			Usage:       "Optional issue number",
		},
	}
}
