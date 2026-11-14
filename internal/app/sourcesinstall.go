package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	shipped "github.com/iandesj/blerg/config"
)

func EnsureDefaultSources(path string) error {
	def, err := DefaultSourcesPath()
	if err != nil {
		return err
	}
	if filepath.Clean(path) != filepath.Clean(def) {
		return nil
	}
	return writeSourcesIfMissing(path, shipped.DefaultYAML)
}

func writeSourcesIfMissing(path string, data []byte) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("sources manifest: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create sources dir: %w", err)
	}
	return createSourcesFile(path, data)
}

func createSourcesFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("create sources manifest: %w", err)
	}
	_, werr := f.Write(data)
	cerr := f.Close()
	if werr != nil {
		return fmt.Errorf("write sources manifest: %w", werr)
	}
	if cerr != nil {
		return fmt.Errorf("write sources manifest: %w", cerr)
	}
	return nil
}
