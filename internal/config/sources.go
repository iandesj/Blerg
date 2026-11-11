package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Source struct {
	Name    string `yaml:"name"`
	FeedURL string `yaml:"feed_url"`
	Enabled bool   `yaml:"enabled"`
}

type SourcesFile struct {
	Sources []Source `yaml:"sources"`
}

func LoadSources(path string) ([]Source, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read sources manifest: %w", err)
	}
	var file SourcesFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse sources manifest: %w", err)
	}
	if len(file.Sources) == 0 {
		return nil, fmt.Errorf("sources manifest %q has no sources", path)
	}
	for i, s := range file.Sources {
		if s.Name == "" {
			return nil, fmt.Errorf("source %d: name is required", i)
		}
		if s.FeedURL == "" {
			return nil, fmt.Errorf("source %q: feed_url is required", s.Name)
		}
	}
	return file.Sources, nil
}
