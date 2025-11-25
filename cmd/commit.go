package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/urfave/cli/v3"

	"github.com/isokolovskii/commitizen/internal/conventional"
)

var ErrInvalidFlag = errors.New("invalid flag")

func commit() *cli.Command {
	var commitType string
	var scope string
	var title string
	var body string
	var footer string

	return &cli.Command{
		Name:  "commit",
		Usage: "Create Conventional Commit",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "type", OnlyOnce: true, Destination: &commitType},
			&cli.StringFlag{Name: "scope", OnlyOnce: true, Destination: &scope},
			&cli.StringFlag{Name: "title", OnlyOnce: true, Destination: &title},
			&cli.StringFlag{Name: "body", OnlyOnce: true, Destination: &body},
			&cli.StringFlag{Name: "footer", OnlyOnce: true, Destination: &footer},
		},
		Action: func(_ context.Context, _ *cli.Command) error {
			if commitType == "" {
				return fmt.Errorf("%w: commitType must be provided", ErrInvalidFlag)
			}
			if title == "" {
				return fmt.Errorf("%w: title mus be provided", ErrInvalidFlag)
			}

			commit, err := conventional.BuildCommit(&conventional.Commit{
				Type:   commitType,
				Scope:  scope,
				Title:  title,
				Body:   body,
				Footer: footer,
			})
			if err != nil {
				return fmt.Errorf("error building commit: %w", err)
			}

			log.Println(commit)

			return nil
		},
	}
}
