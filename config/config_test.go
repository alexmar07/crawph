package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	cfg := Default()
	if cfg.Crawl.MaxDepth != 10 {
		t.Errorf("expected max_depth 10, got %d", cfg.Crawl.MaxDepth)
	}
	if cfg.Crawl.MaxWorkers != 5 {
		t.Errorf("expected max_workers 5, got %d", cfg.Crawl.MaxWorkers)
	}
	if cfg.Crawl.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", cfg.Crawl.Timeout)
	}
	if cfg.Crawl.UserAgent != "Crawph/1.0" {
		t.Errorf("expected user_agent Crawph/1.0, got %s", cfg.Crawl.UserAgent)
	}
	if cfg.RateLimit.DefaultRPS != 1.0 {
		t.Errorf("expected default_rps 1.0, got %f", cfg.RateLimit.DefaultRPS)
	}
	if !cfg.RateLimit.RespectCrawlDelay {
		t.Error("expected respect_crawl_delay true")
	}
	if !cfg.Robots.Enabled {
		t.Error("expected robots enabled")
	}
	if cfg.Storage.Format != "json" {
		t.Errorf("expected format json, got %s", cfg.Storage.Format)
	}
}

func TestLoadFromYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "crawph.yml")
	yaml := `
seeds:
  - https://example.com
  - https://other.com
crawl:
  max_depth: 20
  max_workers: 10
  timeout: 60s
  user_agent: "MyBot/2.0"
rate_limit:
  default_rps: 2.0
  respect_crawl_delay: false
robots:
  enabled: false
storage:
  format: binary
  output: out/data
`
	os.WriteFile(path, []byte(yaml), 0644)
	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Seeds) != 2 {
		t.Errorf("expected 2 seeds, got %d", len(cfg.Seeds))
	}
	if cfg.Crawl.MaxDepth != 20 {
		t.Errorf("expected max_depth 20, got %d", cfg.Crawl.MaxDepth)
	}
	if cfg.Crawl.Timeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", cfg.Crawl.Timeout)
	}
	if cfg.Storage.Format != "binary" {
		t.Errorf("expected format binary, got %s", cfg.Storage.Format)
	}
}

func TestLoadFromFileNotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent.yml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestValidateRejectsNoSeeds(t *testing.T) {
	cfg := Default()
	cfg.Seeds = nil
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty seeds")
	}
}

func TestValidateRejectsInvalidWorkers(t *testing.T) {
	cfg := Default()
	cfg.Seeds = []string{"https://example.com"}
	cfg.Crawl.MaxWorkers = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for 0 workers")
	}
}

func TestValidateRejectsNegativeDepth(t *testing.T) {
	cfg := Default()
	cfg.Seeds = []string{"https://example.com"}
	cfg.Crawl.MaxDepth = -1
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative depth")
	}
}

func TestValidateRejectsZeroTimeout(t *testing.T) {
	cfg := Default()
	cfg.Seeds = []string{"https://example.com"}
	cfg.Crawl.Timeout = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for zero timeout")
	}
}

func TestValidateAcceptsValidConfig(t *testing.T) {
	cfg := Default()
	cfg.Seeds = []string{"https://example.com"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestMergeCLIOverrides(t *testing.T) {
	cfg := Default()
	overrides := CLIOverrides{
		URLs:    "https://a.com,https://b.com",
		Workers: 20,
		Depth:   50,
		Output:  "custom/path",
		Format:  "binary",
		Timeout: 45 * time.Second,
	}
	cfg.ApplyOverrides(overrides)
	if len(cfg.Seeds) != 2 || cfg.Seeds[0] != "https://a.com" {
		t.Errorf("expected overridden seeds, got %v", cfg.Seeds)
	}
	if cfg.Crawl.MaxWorkers != 20 {
		t.Errorf("expected workers 20, got %d", cfg.Crawl.MaxWorkers)
	}
	if cfg.Crawl.MaxDepth != 50 {
		t.Errorf("expected depth 50, got %d", cfg.Crawl.MaxDepth)
	}
	if cfg.Storage.Format != "binary" {
		t.Errorf("expected format binary, got %s", cfg.Storage.Format)
	}
	if cfg.Crawl.Timeout != 45*time.Second {
		t.Errorf("expected timeout 45s, got %v", cfg.Crawl.Timeout)
	}
}

func TestMergeCLIOverridesZeroValuesNotApplied(t *testing.T) {
	cfg := Default()
	cfg.Seeds = []string{"https://example.com"}
	overrides := CLIOverrides{}
	cfg.ApplyOverrides(overrides)
	if cfg.Crawl.MaxWorkers != 5 {
		t.Errorf("expected default workers 5, got %d", cfg.Crawl.MaxWorkers)
	}
}
