package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config is the root configuration.
type Config struct {
	Title      string         `toml:"title"`
	OutputMD   string         `toml:"output_md"`
	OutputHTML string         `toml:"output_html"`
	TokenEnv   string         `toml:"token_env"`
	Sources    []SourceConfig `toml:"sources"`
	Classify   ClassifyConfig `toml:"classify"`
}

// SourceConfig is one GitHub user whose public stars are synced.
type SourceConfig struct {
	Username string `toml:"username"`
	Label    string `toml:"label"`
}

// ClassifyConfig controls dynamic topic-frequency classification.
type ClassifyConfig struct {
	MaxCategories int    `toml:"max_categories"`
	MinCount      int    `toml:"min_count"`
	Fallback      string `toml:"fallback"` // language | other
	SortWithin    string `toml:"sort_within"` // stars | starred_at | name
}

// Load reads and validates a TOML config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Title == "" {
		c.Title = "星标收藏"
	}
	if c.OutputMD == "" && c.OutputHTML == "" {
		c.OutputMD = "STARRED.md"
	}
	if c.TokenEnv == "" {
		c.TokenEnv = "GITHUB_TOKEN"
	}
	if c.Classify.MaxCategories <= 0 {
		c.Classify.MaxCategories = 12
	}
	if c.Classify.MinCount <= 0 {
		c.Classify.MinCount = 2
	}
	fb := strings.ToLower(strings.TrimSpace(c.Classify.Fallback))
	if fb == "" {
		fb = "language"
	}
	c.Classify.Fallback = fb
	sw := strings.ToLower(strings.TrimSpace(c.Classify.SortWithin))
	if sw == "" {
		sw = "stars"
	}
	c.Classify.SortWithin = sw
}

// Validate checks required fields.
func (c *Config) Validate() error {
	if len(c.Sources) == 0 {
		return fmt.Errorf("sources: at least one [[sources]] with username is required")
	}
	for i, s := range c.Sources {
		if strings.TrimSpace(s.Username) == "" {
			return fmt.Errorf("sources[%d].username is required", i)
		}
	}
	switch c.Classify.Fallback {
	case "language", "other":
	default:
		return fmt.Errorf("classify.fallback must be language or other")
	}
	switch c.Classify.SortWithin {
	case "stars", "starred_at", "name":
	default:
		return fmt.Errorf("classify.sort_within must be stars, starred_at, or name")
	}
	if c.OutputMD == "" && c.OutputHTML == "" {
		return fmt.Errorf("at least one of output_md / output_html is required")
	}
	return nil
}

// TokenFromEnv reads an optional token (empty if unset).
func TokenFromEnv(envName string) string {
	return strings.TrimSpace(os.Getenv(envName))
}

// DisplayName returns label or @username.
func (s SourceConfig) DisplayName() string {
	if strings.TrimSpace(s.Label) != "" {
		return strings.TrimSpace(s.Label)
	}
	return "@" + s.Username
}
