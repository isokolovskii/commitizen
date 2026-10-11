// Package retry persists the last generated commit message per repository so
// failed interactive runs can be replayed without answering the wizard again.
package retry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/isokolovskii/commitizen/internal/conventional"
)

type (
	// Store persists generated commits in a JSON file keyed by repository root.
	Store struct {
		// Path is the JSON file backing the store.
		path string
	}
)

const (
	cacheFileName  = "retry.json"
	dirPerm        = 0o750
	filePerm       = 0o600
	readCacheWrap  = "reading retry cache: %w"
	writeCacheWrap = "writing retry cache: %w"
)

// ErrNoCachedCommit reports that no commit was cached for the repository.
var ErrNoCachedCommit = errors.New("no cached commit for repository")

// NewStore returns a Store backed by retry.json inside dir.
func NewStore(dir string) *Store {
	return &Store{path: filepath.Join(dir, cacheFileName)}
}

// Save records the commit generated for the repository rooted at root.
func (store *Store) Save(root string, commit *conventional.Commit) error {
	entries, err := store.read()
	if err != nil {
		return fmt.Errorf(readCacheWrap, err)
	}

	entries[root] = commit

	err = store.write(entries)
	if err != nil {
		return fmt.Errorf(writeCacheWrap, err)
	}

	return nil
}

// Load returns the commit cached for the repository rooted at root.
func (store *Store) Load(root string) (*conventional.Commit, error) {
	entries, err := store.read()
	if err != nil {
		return nil, fmt.Errorf(readCacheWrap, err)
	}

	commit, ok := entries[root]
	if !ok || commit == nil {
		return nil, ErrNoCachedCommit
	}

	return commit, nil
}

// read decodes the cache file, treating a missing file as an empty cache.
func (store *Store) read() (map[string]*conventional.Commit, error) {
	data, err := os.ReadFile(store.path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]*conventional.Commit{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", store.path, err)
	}

	entries := map[string]*conventional.Commit{}

	err = json.Unmarshal(data, &entries)
	if err != nil {
		return nil, fmt.Errorf("decoding %s: %w", store.path, err)
	}

	return entries, nil
}

// write atomically-ish writes the cache file, creating parent directories as needed.
func (store *Store) write(entries map[string]*conventional.Commit) error {
	err := os.MkdirAll(filepath.Dir(store.path), dirPerm)
	if err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding retry cache: %w", err)
	}

	err = os.WriteFile(store.path, data, filePerm)
	if err != nil {
		return fmt.Errorf("writing %s: %w", store.path, err)
	}

	return nil
}
