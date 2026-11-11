package app

import (
	"fmt"
	"os"
	"path/filepath"
)

const DefaultUserAgent = "Blerg (https://github.com/iandesj/Blerg)"

func DefaultDBPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(dir, "Blerg", "blerg.db"), nil
}
