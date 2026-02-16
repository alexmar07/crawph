package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/alexmar07/crawler-go/config"
	"github.com/alexmar07/crawler-go/crawph"
	"github.com/alexmar07/crawler-go/extractor"
	"github.com/alexmar07/crawler-go/graph"
	"github.com/alexmar07/crawler-go/robots"
)

func main() {
	var (
		urls       string
		configFile string
		workers    int
		depth      int
		output     string
		format     string
		timeout    time.Duration
	)

	flag.StringVar(&urls, "urls", "", "Comma-separated list of seed URLs")
	flag.StringVar(&configFile, "config", "", "Path to config file (YAML)")
	flag.IntVar(&workers, "workers", 0, "Number of concurrent workers")
	flag.IntVar(&depth, "depth", 0, "Maximum crawl depth")
	flag.StringVar(&output, "output", "", "Output file path")
	flag.StringVar(&format, "format", "", "Output format (json|binary)")
	flag.DurationVar(&timeout, "timeout", 0, "HTTP request timeout")
	flag.Parse()

	var cfg *config.Config
	var err error

	if configFile != "" {
		cfg, err = config.LoadFromFile(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg = config.Default()
	}

	cfg.ApplyOverrides(config.CLIOverrides{
		URLs:    urls,
		Workers: workers,
		Depth:   depth,
		Output:  output,
		Format:  format,
		Timeout: timeout,
	})

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	var robotsChecker robots.RobotsChecker
	if cfg.Robots.Enabled {
		robotsChecker = robots.NewChecker(cfg.Crawl.UserAgent, cfg.Crawl.Timeout)
	} else {
		robotsChecker = robots.NewDisabledChecker()
	}

	crawler := crawph.New(crawph.Options{
		MaxWorkers:    cfg.Crawl.MaxWorkers,
		MaxDepth:      cfg.Crawl.MaxDepth,
		Timeout:       cfg.Crawl.Timeout,
		UserAgent:     cfg.Crawl.UserAgent,
		DefaultRPS:    cfg.RateLimit.DefaultRPS,
		RobotsChecker: robotsChecker,
		Extractors:    []extractor.Extractor{&extractor.LinkExtractor{}},
		Logger:        logger,
	})

	logger.Info("crawph starting",
		"seeds", cfg.Seeds,
		"workers", cfg.Crawl.MaxWorkers,
		"max_depth", cfg.Crawl.MaxDepth,
	)

	crawler.Start(cfg.Seeds)

	ext := "." + cfg.Storage.Format
	if cfg.Storage.Format == "binary" {
		ext = ".gob"
	}
	outputPath := cfg.Storage.Output + ext

	storage := graph.NewStorage(outputPath, cfg.Storage.Format)
	if err := storage.Save(crawler.Graph()); err != nil {
		logger.Error("failed to save results", "error", err)
		os.Exit(1)
	}

	logger.Info("results saved", "path", outputPath)
}
