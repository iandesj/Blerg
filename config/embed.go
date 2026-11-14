// Package config embeds the shipped sources manifest.
package config

import _ "embed"

//go:embed sources.yaml
var DefaultYAML []byte
