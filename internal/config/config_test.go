package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isokolovskii/commitizen/internal/config"
)

const (
	czrcCustom = `{"path": "custom"}`
	pkgCustom  = `{"config": {"commitizen": {"path": "from-package"}}}`
	czrcHome   = `{"path": "from-home"}`

	czrcFileName    = ".czrc"
	packageJSONName = "package.json"
	pathCustom      = "custom"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()

	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}

func TestDefault(t *testing.T) {
	t.Parallel()

	cfg := config.Default()

	assert.Equal(t, config.DefaultAdapter, cfg.Path)
	assert.Equal(t, config.DefaultMaxSubjectWidth, cfg.MaxSubjectWidth)
	assert.True(t, cfg.ConfirmEnabled())
}

func TestLoadNoFilesYieldsDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(t.TempDir(), t.TempDir())

	require.NoError(t, err)
	assert.Equal(t, config.Default(), cfg)
}

func TestLoadCzrcWinsOverPackageJSONAndHome(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := t.TempDir()
	writeFile(t, root, czrcFileName, czrcCustom)
	writeFile(t, root, packageJSONName, pkgCustom)
	writeFile(t, home, czrcFileName, czrcHome)

	cfg, err := config.Load(root, home)

	require.NoError(t, err)
	assert.Equal(t, pathCustom, cfg.Path)
}

func TestLoadFallsBackToPackageJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	home := t.TempDir()
	writeFile(t, root, packageJSONName, pkgCustom)
	writeFile(t, home, czrcFileName, czrcHome)

	cfg, err := config.Load(root, home)

	require.NoError(t, err)
	assert.Equal(t, "from-package", cfg.Path)
}

func TestLoadFallsBackToHomeCzrc(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(t.TempDir(), writeHomeWithCzrc(t))

	require.NoError(t, err)
	assert.Equal(t, "from-home", cfg.Path)
}

func writeHomeWithCzrc(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	writeFile(t, home, czrcFileName, czrcHome)

	return home
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, root, czrcFileName, `{"path": "custom", "unknown": true}`)

	_, err := config.Load(root, t.TempDir())

	require.ErrorIs(t, err, config.ErrInvalidConfig)
	assert.ErrorContains(t, err, "unknown field")
}

func TestLoadRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, root, czrcFileName, "{not json")

	_, err := config.Load(root, t.TempDir())

	require.ErrorIs(t, err, config.ErrInvalidConfig)
}

func TestLoadWrapsUnreadableFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, czrcFileName), 0o750))

	_, err := config.Load(root, t.TempDir())

	require.ErrorIs(t, err, config.ErrFileNotFound)
}

func TestLoadValidatesValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "negative max subject width",
			content: `{"maxSubjectWidth": -1}`,
			wantErr: "maxSubjectWidth",
		},
		{
			name:    "type without value",
			content: `{"types": [{"value": "", "description": "d"}]}`,
			wantErr: "value must not be empty",
		},
		{
			name:    "type without description",
			content: `{"types": [{"value": "feat"}]}`,
			wantErr: "description must not be empty",
		},
		{
			name:    "duplicate type values",
			content: `{"types": [{"value": "feat", "description": "a"}, {"value": "feat", "description": "b"}]}`,
			wantErr: "duplicate type value",
		},
		{
			name:    "unknown skip question",
			content: `{"skipQuestions": ["wat"]}`,
			wantErr: "unknown skipQuestions entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			writeFile(t, root, czrcFileName, tt.content)

			_, err := config.Load(root, t.TempDir())

			require.ErrorIs(t, err, config.ErrInvalidConfig)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestLoadAcceptsOverrides(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	content := `{"path": "custom", "maxSubjectWidth": 72, "confirm": false, "defaultType": "fix", "skipQuestions": ["scope"]}`
	writeFile(t, root, czrcFileName, content)

	cfg, err := config.Load(root, t.TempDir())

	require.NoError(t, err)
	assert.Equal(t, pathCustom, cfg.Path)
	assert.Equal(t, 72, cfg.MaxSubjectWidth)
	assert.Equal(t, "fix", cfg.DefaultType)
	assert.Equal(t, []string{"scope"}, cfg.SkipQuestions)
	require.NotNil(t, cfg.Confirm)
	assert.False(t, *cfg.Confirm)
	assert.False(t, cfg.ConfirmEnabled())
}

func TestConfirmEnabledDefaultsToTrue(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}

	assert.True(t, cfg.ConfirmEnabled())
}

func TestValidateRejectsCustomTypesWithoutAdapterDefaults(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.Types = []config.Type{{Value: "feat", Description: "features"}}

	require.NoError(t, cfg.Validate())

	cfg.Types = append(cfg.Types, config.Type{Value: "", Description: "empty"})

	err := cfg.Validate()

	require.Error(t, err)
	assert.ErrorIs(t, err, config.ErrInvalidConfig)
}
