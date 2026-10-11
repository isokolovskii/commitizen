package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/isokolovskii/commitizen/internal/adapters"
	conventionaladapter "github.com/isokolovskii/commitizen/internal/adapters/conventional"
	"github.com/isokolovskii/commitizen/internal/config"
	stdconventional "github.com/isokolovskii/commitizen/internal/conventional"
	"github.com/isokolovskii/commitizen/internal/git"
	"github.com/isokolovskii/commitizen/internal/prompts"
	"github.com/isokolovskii/commitizen/internal/retry"
)

type (
	// CommitDeps carries every external interaction of the commit command so
	// tests can substitute fakes.
	CommitDeps struct {
		// Runner performs the git operations.
		Runner git.Runner
		// Prompter drives the interactive wizard.
		Prompter prompts.Prompter
		// Stdout receives normal output.
		Stdout io.Writer
		// Stderr receives warnings.
		Stderr io.Writer
		// Root is the working directory git runs in.
		Root string
		// Home is the user home directory used for fallback configuration.
		Home string
		// CacheDir is the directory backing the retry cache.
		CacheDir string
	}

	// CommitOptions mirrors the commit command flags.
	CommitOptions struct {
		// Commit holds the flag provided commit parts.
		Commit stdconventional.Commit
		// HookFile names a file receiving the message in hook mode.
		HookFile string
		// TypeSet records whether the type flag was given.
		TypeSet bool
		// TitleSet records whether the title flag was given.
		TitleSet bool
		// DryRun prints the message without creating a commit.
		DryRun bool
		// Retry replays the cached commit of this repository.
		Retry bool
		// RetryWithoutEditor is an alias of retry kept for cz compatibility.
		RetryWithoutEditor bool
	}
)

const (
	appName       = "commitizen"
	commitCommand = "commit"

	flagType               = "type"
	flagScope              = "scope"
	flagTitle              = "title"
	flagBody               = "body"
	flagBreaking           = "breaking"
	flagIssue              = "issue"
	flagDryRun             = "dry-run"
	flagHook               = "hook"
	flagRetry              = "retry"
	flagRetryWithoutEditor = "retry-without-editor"

	hookFilePerm       = 0o600
	errBuildingMessage = "building commit message: %w"
	errPrintingMessage = "printing commit message: %w"
	errCreatingCommit  = "creating commit: %w"
	errResolveRoot     = "resolving repository root: %w"
	errBuildRegistry   = "building adapter registry: %w"
)

// ErrNothingStaged is returned when the repository has no staged changes.
var ErrNothingStaged = errors.New("nothing staged to commit")

func commit() *cli.Command {
	return &cli.Command{
		Name:  commitCommand,
		Usage: "create a conventional commit, interactively by default",
		Description: "Without flags the interactive wizard asks type, scope, subject, body, " +
			"breaking changes and issues, then creates the commit.",
		Flags: commitFlags(),
		Action: func(_ context.Context, cmd *cli.Command) error {
			err := runCommit(productionDeps(cmd), optionsFromFlags(cmd))
			if err != nil {
				return mapCommitError(err)
			}

			return nil
		},
	}
}

// runCommit resolves the message, then commits, prints or writes it.
func runCommit(deps *CommitDeps, options *CommitOptions) error {
	message, err := resolveMessage(deps, options)
	if err != nil {
		return fmt.Errorf("resolving message: %w", err)
	}

	err = deliverMessage(deps, options, message)
	if err != nil {
		return fmt.Errorf("delivering message: %w", err)
	}

	return nil
}

// deliverMessage prints, stores or commits the resolved message.
func deliverMessage(deps *CommitDeps, options *CommitOptions, message string) error {
	if options.DryRun {
		err := printCommitMessage(deps.Stdout, message)
		if err != nil {
			return fmt.Errorf("printing dry run message: %w", err)
		}

		return nil
	}

	if options.HookFile != "" {
		err := writeHookFile(deps, options.HookFile, message)
		if err != nil {
			return fmt.Errorf("storing hook message: %w", err)
		}

		return nil
	}

	err := createCommit(deps, message)
	if err != nil {
		return fmt.Errorf("storing commit: %w", err)
	}

	return nil
}

// createCommit stores the commit and prints the resulting message.
func createCommit(deps *CommitDeps, message string) error {
	err := deps.Runner.Commit(message)
	if err != nil {
		return fmt.Errorf(errCreatingCommit, err)
	}

	err = printCommitMessage(deps.Stdout, message)
	if err != nil {
		return fmt.Errorf("printing commit result: %w", err)
	}

	return nil
}

// printCommitMessage writes the final message, which already ends in a newline.
func printCommitMessage(out io.Writer, message string) error {
	_, err := fmt.Fprint(out, message)
	if err != nil {
		return fmt.Errorf(errPrintingMessage, err)
	}

	return nil
}

// resolveMessage picks the flags, retry or wizard path.
func resolveMessage(deps *CommitDeps, options *CommitOptions) (string, error) {
	var (
		message string
		err     error
	)

	switch {
	case options.Retry || options.RetryWithoutEditor:
		message, err = retryMessage(deps)
	case options.TypeSet || options.TitleSet:
		message, err = flagsMessage(options)
	default:
		message, err = wizardMessage(deps)
	}

	if err != nil {
		return "", fmt.Errorf("resolving commit message: %w", err)
	}

	return message, nil
}

// flagsMessage builds the message from the provided flags.
func flagsMessage(options *CommitOptions) (string, error) {
	message, err := stdconventional.BuildCommitMessage(&options.Commit)
	if err != nil {
		return "", fmt.Errorf(errBuildingMessage, err)
	}

	return message, nil
}

// retryMessage replays the cached commit of the current repository.
func retryMessage(deps *CommitDeps) (string, error) {
	root, err := deps.Runner.Root()
	if err != nil {
		return "", fmt.Errorf(errResolveRoot, err)
	}

	commit, err := retry.NewStore(deps.CacheDir).Load(root)
	if err != nil {
		return "", fmt.Errorf("loading retry cache: %w", err)
	}

	message, err := stdconventional.BuildCommitMessage(commit)
	if err != nil {
		return "", fmt.Errorf(errBuildingMessage, err)
	}

	return message, nil
}

// wizardMessage loads configuration, verifies staged files and runs the adapter wizard.
func wizardMessage(deps *CommitDeps) (string, error) {
	root, err := deps.Runner.Root()
	if err != nil {
		return "", fmt.Errorf(errResolveRoot, err)
	}

	cfg, err := config.Load(root, deps.Home)
	if err != nil {
		return "", fmt.Errorf("loading configuration: %w", err)
	}

	err = ensureStaged(deps.Runner)
	if err != nil {
		return "", fmt.Errorf("ensuring staged files: %w", err)
	}

	registry, err := builtinRegistry(cfg)
	if err != nil {
		return "", fmt.Errorf("resolving adapters: %w", err)
	}

	adapter, err := registry.Get(cfg.Path)
	if err != nil {
		return "", fmt.Errorf("resolving adapter: %w", err)
	}

	commit, err := adapter.Prompt(deps.Prompter)
	if err != nil {
		return "", fmt.Errorf("running wizard: %w", err)
	}

	saveRetry(deps, root, commit)

	message, err := stdconventional.BuildCommitMessage(commit)
	if err != nil {
		return "", fmt.Errorf(errBuildingMessage, err)
	}

	return message, nil
}

// builtinRegistry registers every adapter shipped with the binary, honoring cfg.
func builtinRegistry(cfg *config.Config) (*adapters.Registry, error) {
	registry, err := adapters.NewRegistry(conventionaladapter.New(cfg))
	if err != nil {
		return nil, fmt.Errorf(errBuildRegistry, err)
	}

	return registry, nil
}

// ensureStaged fails when there is nothing to commit.
func ensureStaged(runner git.Runner) error {
	staged, err := runner.StagedFiles()
	if err != nil {
		return fmt.Errorf("listing staged files: %w", err)
	}

	if len(staged) == 0 {
		return ErrNothingStaged
	}

	return nil
}

// saveRetry caches the generated commit, warning instead of failing.
func saveRetry(deps *CommitDeps, root string, commit *stdconventional.Commit) {
	err := retry.NewStore(deps.CacheDir).Save(root, commit)
	if err != nil {
		_, _ = fmt.Fprintf(deps.Stderr, "warning: could not save retry cache: %s\n", err)
	}
}

// productionDeps wires the command to the real git binary, terminal and cache.
func productionDeps(cmd *cli.Command) *CommitDeps {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	return &CommitDeps{
		Runner:   git.NewRunner(wd),
		Prompter: prompts.Interactive{},
		Stdout:   cmd.Root().Writer,
		Stderr:   os.Stderr,
		Root:     wd,
		Home:     home,
		CacheDir: userCacheDir(),
	}
}

// userCacheDir returns the commitizen cache directory, falling back to the
// system temporary directory.
func userCacheDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}

	return filepath.Join(base, appName)
}

// writeHookFile stores the message for the pending prepare-commit-msg hook run.
func writeHookFile(deps *CommitDeps, path, message string) error {
	err := os.WriteFile(path, []byte(message), hookFilePerm)
	if err != nil {
		return fmt.Errorf("writing hook file %s: %w", path, err)
	}

	_, err = fmt.Fprintf(deps.Stdout, "Wrote commit message to %s\n", path)
	if err != nil {
		return fmt.Errorf("printing hook confirmation: %w", err)
	}

	return nil
}

// mapCommitError converts known failures into friendly CLI exits.
func mapCommitError(err error) error {
	switch {
	case errors.Is(err, git.ErrNotRepository):
		return cli.Exit("not inside a git repository", 1)
	case errors.Is(err, ErrNothingStaged):
		return cli.Exit("nothing staged: stage your changes before running commitizen", 1)
	case errors.Is(err, retry.ErrNoCachedCommit):
		return cli.Exit("no cached commit to retry, run commitizen interactively first", 1)
	case errors.Is(err, adapters.ErrUnknownAdapter):
		return cli.Exit(err.Error(), 1)
	case errors.Is(err, conventionaladapter.ErrCommitCanceled):
		return cli.Exit("commit canceled", 1)
	default:
		return cli.Exit(err.Error(), 1)
	}
}

// optionsFromFlags reads the commit command flags into options.
func optionsFromFlags(cmd *cli.Command) *CommitOptions {
	return &CommitOptions{
		Commit: stdconventional.Commit{
			Type:           cmd.String(flagType),
			Scope:          cmd.String(flagScope),
			Title:          cmd.String(flagTitle),
			Body:           cmd.String(flagBody),
			BreakingChange: cmd.String(flagBreaking),
			Issue:          cmd.String(flagIssue),
		},
		HookFile:           cmd.String(flagHook),
		TypeSet:            cmd.IsSet(flagType),
		TitleSet:           cmd.IsSet(flagTitle),
		DryRun:             cmd.Bool(flagDryRun),
		Retry:              cmd.Bool(flagRetry),
		RetryWithoutEditor: cmd.Bool(flagRetryWithoutEditor),
	}
}

// commitFlags declares every flag of the commit command.
func commitFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: flagType, Usage: "commit type, for example feat or fix"},
		&cli.StringFlag{Name: flagScope, Usage: "optional scope of the change"},
		&cli.StringFlag{Name: flagTitle, Usage: "short imperative summary"},
		&cli.StringFlag{Name: flagBody, Usage: "longer description"},
		&cli.BoolFlag{Name: flagBreaking, Usage: "mark the commit as breaking"},
		&cli.StringFlag{Name: flagIssue, Usage: "issue reference, for example JIRA-1"},
		&cli.BoolFlag{Name: flagDryRun, Usage: "print the message instead of committing"},
		&cli.StringFlag{Name: flagHook, Usage: "write the message to the given file, for prepare-commit-msg hooks"},
		&cli.BoolFlag{Name: flagRetry, Usage: "repeat the last commit of this repository"},
		&cli.BoolFlag{Name: flagRetryWithoutEditor, Usage: "alias of retry kept for cz compatibility"},
	}
}
