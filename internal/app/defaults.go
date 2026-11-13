package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const DefaultUserAgent = "Blerg (https://github.com/iandesj/Blerg)"

const (
	DefaultStaleness     = 24 * time.Hour
	DefaultMaxConcurrent = 4
)

func DefaultDBPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(dir, "Blerg", "blerg.db"), nil
}
