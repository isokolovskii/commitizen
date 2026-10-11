package adapters_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isokolovskii/commitizen/internal/adapters"
	"github.com/isokolovskii/commitizen/internal/conventional"
	"github.com/isokolovskii/commitizen/internal/prompts"
)

type (
	fakeAdapter struct {
		// Name is the registry key returned by Name.
		name string
	}
)

const (
	adapterNameAlpha = "alpha"
	adapterNameBeta  = "beta"
)

func (fake *fakeAdapter) Name() string {
	return fake.name
}

func (fake *fakeAdapter) Description() string {
	return "adapter " + fake.name
}

func (fake *fakeAdapter) Example() (string, error) {
	return "feat: " + fake.name, nil
}

func (fake *fakeAdapter) Prompt(_ prompts.Prompter) (*conventional.Commit, error) {
	return &conventional.Commit{}, nil
}

func newTestRegistry(t *testing.T) *adapters.Registry {
	t.Helper()

	registry, err := adapters.NewRegistry(&fakeAdapter{name: adapterNameBeta}, &fakeAdapter{name: adapterNameAlpha})
	require.NoError(t, err)

	return registry
}

func TestNewRegistrySortsNames(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)

	assert.Equal(t, []string{adapterNameAlpha, adapterNameBeta}, registry.Names())
}

func TestNewRegistryRejectsEmptyName(t *testing.T) {
	t.Parallel()

	_, err := adapters.NewRegistry(&fakeAdapter{name: ""})

	require.ErrorIs(t, err, adapters.ErrEmptyName)
}

func TestNewRegistryRejectsDuplicates(t *testing.T) {
	t.Parallel()

	_, err := adapters.NewRegistry(&fakeAdapter{name: "x"}, &fakeAdapter{name: "x"})

	require.ErrorIs(t, err, adapters.ErrDuplicateAdapter)
}

func TestGetReturnsRegisteredAdapter(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)

	adapter, err := registry.Get(adapterNameAlpha)

	require.NoError(t, err)
	assert.Equal(t, adapterNameAlpha, adapter.Name())
	assert.Equal(t, "adapter "+adapterNameAlpha, adapter.Description())

	message, err := adapter.Example()

	require.NoError(t, err)
	assert.Equal(t, "feat: "+adapterNameAlpha, message)
}

func TestGetUnknownAdapterListsAvailable(t *testing.T) {
	t.Parallel()

	registry := newTestRegistry(t)

	_, err := registry.Get("missing")

	require.Error(t, err)
	require.ErrorIs(t, err, adapters.ErrUnknownAdapter)
	assert.ErrorContains(t, err, "available: ["+adapterNameAlpha+" "+adapterNameBeta+"]")
}
