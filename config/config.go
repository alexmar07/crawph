package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Seeds     []string        `yaml:"seeds"`
	Crawl     CrawlConfig     `yaml:"crawl"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Robots    RobotsConfig    `yaml:"robots"`
	Storage   StorageConfig   `yaml:"storage"`
}

type CrawlConfig struct {
	MaxDepth   int           `yaml:"max_depth"`
	MaxWorkers int           `yaml:"max_workers"`
	Timeout    time.Duration `yaml:"timeout"`
	UserAgent  string        `yaml:"user_agent"`
}

type RateLimitConfig struct {
	DefaultRPS        float64 `yaml:"default_rps"`
	RespectCrawlDelay bool    `yaml:"respect_crawl_delay"`
}

type RobotsConfig struct {
	Enabled bool `yaml:"enabled"`
}

type StorageConfig struct {
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type CLIOverrides struct {
	URLs    string
	Workers int
	Depth   int
	Output  string
	Format  string
	Timeout time.Duration
}

func Default() *Config {
	return &Config{
		Crawl: CrawlConfig{
			MaxDepth:   10,
			MaxWorkers: 5,
			Timeout:    30 * time.Second,
			UserAgent:  "Crawph/1.0",
		},
		RateLimit: RateLimitConfig{
			DefaultRPS:        1.0,
			RespectCrawlDelay: true,
		},
		Robots: RobotsConfig{
			Enabled: true,
		},
		Storage: StorageConfig{
			Format: "json",
			Output: "data/result",
		},
	}
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}
	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Seeds) == 0 {
		return fmt.Errorf("at least one seed URL is required")
	}
	if c.Crawl.MaxWorkers < 1 {
		return fmt.Errorf("max_workers must be at least 1, got %d", c.Crawl.MaxWorkers)
	}
	if c.Crawl.MaxDepth < 0 {
		return fmt.Errorf("max_depth must be non-negative, got %d", c.Crawl.MaxDepth)
	}
	if c.Crawl.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %v", c.Crawl.Timeout)
	}
	return nil
}

func (c *Config) ApplyOverrides(o CLIOverrides) {
	if o.URLs != "" {
		c.Seeds = strings.Split(o.URLs, ",")
	}
	if o.Workers > 0 {
		c.Crawl.MaxWorkers = o.Workers
	}
	if o.Depth > 0 {
		c.Crawl.MaxDepth = o.Depth
	}
	if o.Output != "" {
		c.Storage.Output = o.Output
	}
	if o.Format != "" {
		c.Storage.Format = o.Format
	}
	if o.Timeout > 0 {
		c.Crawl.Timeout = o.Timeout
	}
}
