package conventional

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/isokolovskii/commitizen/internal/config"
	"github.com/isokolovskii/commitizen/internal/conventional"
	"github.com/isokolovskii/commitizen/internal/prompts"
)

type (
	// StepFunc is one wizard question mutating the commit under construction.
	stepFunc func(*conventional.Commit, prompts.Prompter) error
)

const (
	selectTypeMessage           = "Select the type of change that you're committing"
	askScopeMessage             = "What is the scope of this change (e.g. component or file name)"
	askSubjectMessage           = "Write a short, imperative tense description of the change"
	askBodyMessage              = "Provide a longer description of the change"
	askBreakingMessage          = "Are there any breaking changes?"
	askBreakingDescriptionLabel = "Describe the breaking changes"
	askIssuesMessage            = "List any issue numbers closed by this change (e.g. #123)"
	askProceedMessage           = "Proceed with this commit message?"
)

// Prompt runs the interactive wizard and returns the generated commit.
func (adapter *Adapter) Prompt(prompter prompts.Prompter) (*conventional.Commit, error) {
	commit := &conventional.Commit{}

	for _, step := range adapter.steps() {
		err := step(commit, prompter)
		if err != nil {
			return nil, fmt.Errorf("running wizard step: %w", err)
		}
	}

	return commit, nil
}

// steps orders the wizard questions, honoring the skipQuestions list.
func (adapter *Adapter) steps() []stepFunc {
	return []stepFunc{
		adapter.askTypeStep(),
		adapter.skipStep(config.QuestionScope, adapter.askScopeStep),
		adapter.askSubjectStep(),
		adapter.skipStep(config.QuestionBody, adapter.askBodyStep),
		adapter.skipStep(config.QuestionBreaking, adapter.askBreakingStep),
		adapter.skipStep(config.QuestionIssues, adapter.askIssuesStep),
		adapter.skipStep(config.QuestionConfirm, adapter.confirmStep()),
	}
}

// skipStep neutralizes the next question when it is listed in skipQuestions.
func (adapter *Adapter) skipStep(question string, next stepFunc) stepFunc {
	if slices.Contains(adapter.cfg.SkipQuestions, question) {
		return func(*conventional.Commit, prompts.Prompter) error {
			return nil
		}
	}

	return next
}

// askTypeStep returns the wizard step selecting the commit type.
func (adapter *Adapter) askTypeStep() func(*conventional.Commit, prompts.Prompter) error {
	options := adapter.typeOptions()
	defaultIndex := defaultTypeIndex(adapter.cfg, options)

	return func(commit *conventional.Commit, prompter prompts.Prompter) error {
		index, err := prompter.Select(selectTypeMessage, options, defaultIndex)
		if err != nil {
			return fmt.Errorf("selecting type: %w", err)
		}

		commit.Type = options[index].Value

		return nil
	}
}

// askScopeStep asks for the optional scope of the change.
func (adapter *Adapter) askScopeStep(commit *conventional.Commit, prompter prompts.Prompter) error {
	scope, err := prompter.Input(askScopeMessage, adapter.cfg.DefaultScope, nil)
	if err != nil {
		return fmt.Errorf("asking for scope: %w", err)
	}

	commit.Scope = scope

	return nil
}

// askSubjectStep asks for the subject, validating it against the configured width.
func (adapter *Adapter) askSubjectStep() func(*conventional.Commit, prompts.Prompter) error {
	return func(commit *conventional.Commit, prompter prompts.Prompter) error {
		subject, err := prompter.Input(askSubjectMessage, adapter.cfg.DefaultSubject, adapter.validateSubject)
		if err != nil {
			return fmt.Errorf("asking for subject: %w", err)
		}

		commit.Title = subject

		return nil
	}
}

// askBodyStep asks for the optional longer description of the change.
func (adapter *Adapter) askBodyStep(commit *conventional.Commit, prompter prompts.Prompter) error {
	body, err := prompter.MultiLine(askBodyMessage, adapter.cfg.DefaultBody)
	if err != nil {
		return fmt.Errorf("asking for body: %w", err)
	}

	commit.Body = body

	return nil
}

// askBreakingStep asks whether the change breaks anything and, if so, why.
func (*Adapter) askBreakingStep(commit *conventional.Commit, prompter prompts.Prompter) error {
	breaking, err := prompter.Confirm(askBreakingMessage, prompts.DefaultNo)
	if err != nil {
		return fmt.Errorf("asking about breaking changes: %w", err)
	}

	if !breaking {
		return nil
	}

	description, err := prompter.Input(askBreakingDescriptionLabel, "", nil)
	if err != nil {
		return fmt.Errorf("asking for breaking change description: %w", err)
	}

	commit.BreakingChange = description

	return nil
}

// askIssuesStep asks for the issue references closed by the change.
func (adapter *Adapter) askIssuesStep(commit *conventional.Commit, prompter prompts.Prompter) error {
	issue, err := prompter.Input(askIssuesMessage, adapter.cfg.DefaultIssues, nil)
	if err != nil {
		return fmt.Errorf("asking for issues: %w", err)
	}

	commit.Issue = issue

	return nil
}

// confirmStep asks the user to accept the fully assembled message.
func (adapter *Adapter) confirmStep() func(*conventional.Commit, prompts.Prompter) error {
	if !adapter.cfg.ConfirmEnabled() {
		return func(*conventional.Commit, prompts.Prompter) error {
			return nil
		}
	}

	return func(_ *conventional.Commit, prompter prompts.Prompter) error {
		proceed, err := prompter.Confirm(askProceedMessage, prompts.DefaultYes)
		if err != nil {
			return fmt.Errorf("asking to proceed: %w", err)
		}

		if !proceed {
			return ErrCommitCanceled
		}

		return nil
	}
}

// typeOptions converts the resolved type table into select prompt options.
func (adapter *Adapter) typeOptions() []prompts.Option {
	types := Types(adapter.cfg)

	options := make([]prompts.Option, 0, len(types))
	for _, entry := range types {
		options = append(options, prompts.Option{
			Value: entry.Value,
			Label: entry.Value + ": " + entry.Description,
		})
	}

	return options
}

// validateSubject rejects empty subjects and subjects wider than configured.
func (adapter *Adapter) validateSubject(subject string) error {
	if strings.TrimSpace(subject) == "" {
		return fmt.Errorf("%w: subject", conventional.ErrRequiredPartNotPreset)
	}

	width := utf8.RuneCountInString(subject)
	if width > adapter.cfg.MaxSubjectWidth {
		return fmt.Errorf(
			"%w: subject is %d characters, max is %d",
			ErrSubjectTooLong,
			width,
			adapter.cfg.MaxSubjectWidth,
		)
	}

	return nil
}

// defaultTypeIndex resolves the configured default type to an option index,
// falling back to the first option.
func defaultTypeIndex(cfg *config.Config, options []prompts.Option) int {
	index := slices.IndexFunc(options, func(option prompts.Option) bool {
		return option.Value == cfg.DefaultType
	})
	if index < 0 {
		return 0
	}

	return index
}
