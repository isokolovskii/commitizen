package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/isokolovskii/commitizen/internal/adapters"
	"github.com/isokolovskii/commitizen/internal/config"
)

type (
	// CzrcDocument is the JSON shape written to .czrc files.
	CzrcDocument struct {
		// Path is the adapter registry key.
		Path string `json:"path"`
	}

	// InitDeps carries the destinations the init command writes to.
	InitDeps struct {
		// Stdout receives the confirmation output.
		Stdout io.Writer
		// Dir is the directory receiving the .czrc file.
		Dir string
	}
)

const (
	czrcFileName     = ".czrc"
	forceFlag        = "force"
	errPreparingInit = "preparing init: %w"
)

// ErrConfigExists is returned when .czrc exists and --force was not given.
var ErrConfigExists = errors.New("configuration file already exists")

func initCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "initialize .czrc with the given adapter",
		ArgsUsage: "[adapter]",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: forceFlag, Usage: "overwrite an existing .czrc"},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			name := cmd.Args().First()
			if name == "" {
				name = config.DefaultAdapter
			}

			deps := &InitDeps{Stdout: os.Stdout, Dir: "."}

			err := runInit(deps, name, cmd.Bool(forceFlag))
			if err != nil {
				return mapInitError(err)
			}

			return nil
		},
	}
}

// runInit validates the adapter name and writes the .czrc file.
func runInit(deps *InitDeps, name string, force bool) error {
	registry, err := builtinRegistry(config.Default())
	if err != nil {
		return fmt.Errorf("building adapters: %w", err)
	}

	_, err = registry.Get(name)
	if err != nil {
		return fmt.Errorf("validating adapter: %w", err)
	}

	path := filepath.Join(deps.Dir, czrcFileName)

	err = ensureInitTarget(path, force)
	if err != nil {
		return fmt.Errorf(errPreparingInit, err)
	}

	data, err := json.MarshalIndent(CzrcDocument{Path: name}, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding configuration: %w", err)
	}

	err = os.WriteFile(path, append(data, '\n'), hookFilePerm)
	if err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	_, err = fmt.Fprintf(deps.Stdout, "Initialized %s with adapter %s\n", path, name)
	if err != nil {
		return fmt.Errorf("printing confirmation: %w", err)
	}

	return nil
}

// ensureInitTarget fails when the target exists and force was not requested.
func ensureInitTarget(path string, force bool) error {
	_, err := os.Stat(path)
	switch {
	case err == nil && !force:
		return fmt.Errorf("%w: %s, use --force to overwrite", ErrConfigExists, path)
	case err != nil && !errors.Is(err, os.ErrNotExist):
		return fmt.Errorf("checking %s: %w", path, err)
	}

	return nil
}

// mapInitError converts known init failures into CLI exits.
func mapInitError(err error) error {
	switch {
	case errors.Is(err, adapters.ErrUnknownAdapter), errors.Is(err, ErrConfigExists):
		return cli.Exit(err.Error(), 1)
	default:
		return err
	}
}
