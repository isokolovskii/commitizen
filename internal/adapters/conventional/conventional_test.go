package conventional_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isokolovskii/commitizen/internal/adapters/conventional"
	"github.com/isokolovskii/commitizen/internal/config"
)

const (
	testDefaultSubject  = "polish retry cache"
	testFeatValue       = "feat"
	testRevertValue     = "revert"
	testCustomTypeValue = "sparkle"
	testCustomTypeDesc  = "Add glitter"
)

func TestAdapterMetadata(t *testing.T) {
	t.Parallel()

	adapter := conventional.New(config.Default())

	assert.Equal(t, "conventional", adapter.Name())
	assert.NotEmpty(t, adapter.Description())
}

func TestExampleBuildsConventionalMessage(t *testing.T) {
	t.Parallel()

	adapter := conventional.New(config.Default())

	message, err := adapter.Example()

	require.NoError(t, err)
	assert.Equal(
		t,
		"feat(core): add hello world handler\n\nProvide a working baseline for the generated example message.\n",
		message,
	)
}

func TestExampleUsesConfiguredDefaultSubject(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.DefaultSubject = testDefaultSubject

	adapter := conventional.New(cfg)

	message, err := adapter.Example()

	require.NoError(t, err)
	assert.Contains(t, message, testDefaultSubject)
}

func TestTypesDefaultsToBuiltinTable(t *testing.T) {
	t.Parallel()

	types := conventional.Types(nil)

	require.Len(t, types, 11)
	assert.Equal(t, testFeatValue, types[0].Value)
	assert.Equal(t, "A new feature", types[0].Description)
	assert.Equal(t, testRevertValue, types[10].Value)

	values := make([]string, 0, len(types))
	for _, typ := range types {
		values = append(values, typ.Value)
	}

	assert.Equal(
		t,
		[]string{
			testFeatValue,
			"fix",
			"docs",
			"style",
			"refactor",
			"perf",
			"test",
			"build",
			"ci",
			"chore",
			testRevertValue,
		},
		values,
	)
}

func TestTypesPrefersConfigOverrides(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	cfg.Types = []config.Type{{Value: testCustomTypeValue, Description: testCustomTypeDesc}}

	types := conventional.Types(cfg)

	require.Len(t, types, 1)
	assert.Equal(t, testCustomTypeValue, types[0].Value)
	assert.Equal(t, testCustomTypeDesc, types[0].Description)
}
