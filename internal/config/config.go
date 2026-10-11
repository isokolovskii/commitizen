// Package config loads and validates Commitizen adapter configuration from
// repository level and user level files.
package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type (
	// Type describes one selectable commit type of the wizard.
	Type struct {
		// Value is the commit type keyword written into the header.
		Value string `json:"value"`
		// Name is an optional display name, defaults to Value.
		Name string `json:"name,omitempty"`
		// Description explains when to use the type.
		Description string `json:"description"`
	}

	// Config is the Commitizen configuration schema shared by .czrc files
	// and the package.json config.commitizen key. Field order follows the
	// struct alignment advice of fieldalignment.
	Config struct {
		Confirm         *bool    `json:"confirm,omitempty"`
		Path            string   `json:"path"`
		DefaultType     string   `json:"defaultType,omitempty"`
		DefaultScope    string   `json:"defaultScope,omitempty"`
		DefaultSubject  string   `json:"defaultSubject,omitempty"`
		DefaultBody     string   `json:"defaultBody,omitempty"`
		DefaultIssues   string   `json:"defaultIssues,omitempty"`
		Types           []Type   `json:"types,omitempty"`
		SkipQuestions   []string `json:"skipQuestions,omitempty"`
		MaxSubjectWidth int      `json:"maxSubjectWidth,omitempty"`
	}

	// ConfigSource is one candidate configuration file location.
	configSource struct {
		// Path is the absolute file location.
		path string
		// Kind selects the file parser.
		kind string
	}

	// PackageJSONConfig mirrors the config key of package.json.
	packageJSONConfig struct {
		// Commitizen holds the raw config.commitizen payload.
		Commitizen json.RawMessage `json:"commitizen"`
	}

	// PackageJSON mirrors the subset of package.json this CLI reads.
	packageJSON struct {
		// Config holds tool specific configuration.
		Config packageJSONConfig `json:"config"`
	}
)

const (
	// DefaultAdapter is used when no adapter is configured.
	DefaultAdapter = "conventional"
	// DefaultMaxSubjectWidth caps the commit subject length like the
	// conventional changelog ecosystem does.
	DefaultMaxSubjectWidth = 100

	// QuestionScope and the following Question* constants name the entries
	// allowed in the skipQuestions configuration list.
	QuestionScope    = "scope"
	QuestionBody     = "body"
	QuestionBreaking = "breaking"
	QuestionIssues   = "issues"
	QuestionConfirm  = "confirm"

	czrcFileName    = ".czrc"
	packageJSONName = "package.json"
	kindCzrc        = "czrc"
	kindPackageJSON = "package"
	fmtContextWrap  = "%w: %s: %w"
)

var (
	// ErrInvalidConfig marks semantically wrong configuration values.
	ErrInvalidConfig = errors.New("invalid config")
	// ErrFileNotFound is returned with the wrapped os error when a required
	// configuration file cannot be read.
	ErrFileNotFound = errors.New("config file not found")

	errTypeValueEmpty       = errors.New("value must not be empty")
	errTypeDescriptionEmpty = errors.New("description must not be empty")
	errDuplicateTypeValue   = errors.New("duplicate type value")
	errNoCommitizenConfig   = errors.New("no commitizen config in package.json")
)

// Default returns the configuration used when no file provides one.
func Default() *Config {
	confirm := true

	return &Config{
		Path:            DefaultAdapter,
		MaxSubjectWidth: DefaultMaxSubjectWidth,
		Confirm:         &confirm,
	}
}

// Load resolves configuration in cz-cli compatible order, first match wins:
// gitRoot/.czrc, then gitRoot/package.json (config.commitizen key), then
// homeDir/.czrc. Missing sources everywhere yields Default.
func Load(gitRoot, homeDir string) (*Config, error) {
	sources := []configSource{
		{path: filepath.Join(gitRoot, czrcFileName), kind: kindCzrc},
		{path: filepath.Join(gitRoot, packageJSONName), kind: kindPackageJSON},
		{path: filepath.Join(homeDir, czrcFileName), kind: kindCzrc},
	}

	for _, source := range sources {
		cfg, found, err := loadFile(source.path, source.kind)
		if err != nil {
			return nil, fmt.Errorf("resolving config: %w", err)
		}

		if found {
			return cfg, nil
		}
	}

	return Default(), nil
}

// ConfirmEnabled reports whether the final wizard review step is active,
// defaulting to true when the config left it unset.
func (cfg *Config) ConfirmEnabled() bool {
	return cfg.Confirm == nil || *cfg.Confirm
}

// Validate checks semantic rules that defaults cannot repair.
func (cfg *Config) Validate() error {
	if cfg.MaxSubjectWidth < 0 {
		return fmt.Errorf("%w: maxSubjectWidth must not be negative, got %d", ErrInvalidConfig, cfg.MaxSubjectWidth)
	}

	seen := make(map[string]struct{}, len(cfg.Types))
	for i, typ := range cfg.Types {
		err := validateType(typ, seen)
		if err != nil {
			return fmt.Errorf("%w: types[%d] %q: %w", ErrInvalidConfig, i, typ.Value, err)
		}
	}

	for _, question := range cfg.SkipQuestions {
		if !isKnownQuestion(question) {
			return fmt.Errorf(
				"%w: unknown skipQuestions entry %q, allowed: %s, %s, %s, %s, %s",
				ErrInvalidConfig,
				question,
				QuestionScope,
				QuestionBody,
				QuestionBreaking,
				QuestionIssues,
				QuestionConfirm,
			)
		}
	}

	return nil
}

func validateType(typ Type, seen map[string]struct{}) error {
	if typ.Value == "" {
		return errTypeValueEmpty
	}

	if typ.Description == "" {
		return errTypeDescriptionEmpty
	}

	if _, ok := seen[typ.Value]; ok {
		return errDuplicateTypeValue
	}

	seen[typ.Value] = struct{}{}

	return nil
}

func isKnownQuestion(question string) bool {
	switch question {
	case QuestionScope, QuestionBody, QuestionBreaking, QuestionIssues, QuestionConfirm:
		return true
	default:
		return false
	}
}

func loadFile(path, kind string) (*Config, bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}

		return nil, false, fmt.Errorf(fmtContextWrap, ErrFileNotFound, path, err)
	}

	cfg, err := parse(content, kind)
	if err != nil {
		if errors.Is(err, errNoCommitizenConfig) {
			return nil, false, nil
		}

		return nil, false, fmt.Errorf(fmtContextWrap, ErrInvalidConfig, path, err)
	}

	err = cfg.Validate()
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", path, err)
	}

	applyDefaults(cfg)

	return cfg, true, nil
}

func parse(content []byte, kind string) (*Config, error) {
	if kind == kindPackageJSON {
		cfg, err := parsePackageJSON(content)
		if err != nil {
			return nil, fmt.Errorf("reading config.commitizen: %w", err)
		}

		return cfg, nil
	}

	cfg, err := decodeStrict(content)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", czrcFileName, err)
	}

	return cfg, nil
}

func parsePackageJSON(content []byte) (*Config, error) {
	var pkg packageJSON

	err := json.Unmarshal(content, &pkg)
	if err != nil {
		return nil, fmt.Errorf("decoding package.json: %w", err)
	}

	if len(pkg.Config.Commitizen) == 0 {
		return nil, errNoCommitizenConfig
	}

	cfg, err := decodeStrict(pkg.Config.Commitizen)
	if err != nil {
		return nil, fmt.Errorf("decoding config.commitizen: %w", err)
	}

	return cfg, nil
}

func decodeStrict(content []byte) (*Config, error) {
	cfg := &Config{}

	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()

	err := decoder.Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}

	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Path == "" {
		cfg.Path = DefaultAdapter
	}

	if cfg.MaxSubjectWidth == 0 {
		cfg.MaxSubjectWidth = DefaultMaxSubjectWidth
	}

	if cfg.Confirm == nil {
		confirm := true
		cfg.Confirm = &confirm
	}
}
