package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExample(t *testing.T) {
	path := filepath.Join("..", "..", "configs", "config.example.toml")
	if _, err := os.Stat(path); err != nil {
		t.Skip("example config not found")
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Sources) == 0 {
		t.Fatal("expected sources")
	}
	if cfg.Classify.MaxCategories != 12 {
		t.Fatalf("max_categories=%d", cfg.Classify.MaxCategories)
	}
}
