package crawph

import (
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/alexmar07/crawler-go/extractor"
	"github.com/alexmar07/crawler-go/fetcher"
	"github.com/alexmar07/crawler-go/graph"
	"github.com/alexmar07/crawler-go/queue"
	"github.com/alexmar07/crawler-go/ratelimit"
	"github.com/alexmar07/crawler-go/robots"
)

type Crawph struct {
	maxWorkers        int
	maxDepth          int
	respectCrawlDelay bool
	graph             *graph.Graph
	queue             *queue.Queue
	fetcher           *fetcher.Fetcher
	extractors        []extractor.Extractor
	robotsChecker     robots.RobotsChecker
	rateLimiter       *ratelimit.Registry
	visited           sync.Map
	activeWorkers     int
	activeMu          sync.Mutex
	wg                sync.WaitGroup
	logger            *slog.Logger
}

type Options struct {
	MaxWorkers        int
	MaxDepth          int
	Timeout           time.Duration
	UserAgent         string
	DefaultRPS        float64
	RespectCrawlDelay bool
	RobotsChecker     robots.RobotsChecker
	Extractors        []extractor.Extractor
	Logger            *slog.Logger
}

func New(opts Options) *Crawph {
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.RobotsChecker == nil {
		opts.RobotsChecker = robots.NewDisabledChecker()
	}
	if len(opts.Extractors) == 0 {
		opts.Extractors = []extractor.Extractor{&extractor.LinkExtractor{}}
	}

	return &Crawph{
		maxWorkers:        opts.MaxWorkers,
		maxDepth:          opts.MaxDepth,
		respectCrawlDelay: opts.RespectCrawlDelay,
		graph:         graph.NewGraph(),
		queue:         queue.NewQueue(),
		fetcher:       fetcher.New(opts.Timeout, opts.UserAgent, 10),
		extractors:    opts.Extractors,
		robotsChecker: opts.RobotsChecker,
		rateLimiter:   ratelimit.NewRegistry(opts.DefaultRPS),
		logger:        opts.Logger,
	}
}

func (c *Crawph) Graph() *graph.Graph {
	return c.graph
}

func (c *Crawph) Start(seeds []string) {
	for _, seed := range seeds {
		normalized, err := graph.NormalizeURL(seed)
		if err != nil {
			c.logger.Error("invalid seed URL", "url", seed, "error", err)
			continue
		}
		c.visited.Store(normalized, true)
		c.queue.Enqueue(queue.Item{URL: normalized, Depth: 0})
		c.logger.Info("seed enqueued", "url", normalized)
	}

	for i := 0; i < c.maxWorkers; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}

	go c.monitorWorkers()

	c.wg.Wait()
	c.logger.Info("crawl completed", "vertices", c.graph.VertexCount())
}

func (c *Crawph) worker(id int) {
	defer c.wg.Done()

	for {
		item, ok := c.queue.Dequeue()
		if !ok {
			c.logger.Debug("worker terminated", "id", id)
			return
		}

		c.incActive()

		if item.Depth > c.maxDepth {
			c.logger.Debug("max depth reached", "url", item.URL, "depth", item.Depth)
			c.decActive()
			continue
		}

		c.processURL(item)
		c.decActive()
	}
}

func (c *Crawph) processURL(item queue.Item) {
	// Check robots.txt
	allowed, err := c.robotsChecker.IsAllowed(item.URL)
	if err != nil {
		c.logger.Error("robots check failed", "url", item.URL, "error", err)
	}
	if !allowed {
		c.logger.Debug("blocked by robots.txt", "url", item.URL)
		return
	}

	// Rate limit (apply crawl-delay override if configured)
	domain := extractDomain(item.URL)
	crawlDelay := c.robotsChecker.GetCrawlDelay(item.URL)
	if c.respectCrawlDelay && crawlDelay > 0 {
		rps := 1.0 / crawlDelay.Seconds()
		c.rateLimiter.SetDomainRate(domain, rps)
	}
	c.rateLimiter.GetLimiter(domain).Wait()

	// Fetch
	doc, err := c.fetcher.Fetch(item.URL)
	if err != nil {
		c.logger.Error("fetch failed", "url", item.URL, "error", err)
		return
	}

	// Add source vertex (item.URL is already normalized)
	sourceVertex, err := c.graph.AddVertexNormalized(item.URL)
	if err != nil {
		c.logger.Error("add vertex failed", "url", item.URL, "error", err)
		return
	}

	// Run extractors
	for _, ext := range c.extractors {
		result, err := ext.Extract(doc, item.URL)
		if err != nil {
			c.logger.Error("extractor failed",
				"extractor", ext.Name(), "url", item.URL, "error", err)
			continue
		}

		for _, link := range result.Links {
			normalized, err := graph.NormalizeURL(link)
			if err != nil {
				continue
			}

			targetVertex, err := c.graph.AddVertexNormalized(normalized)
			if err != nil {
				continue
			}
			c.graph.AddEdge(sourceVertex, targetVertex)

			if _, loaded := c.visited.LoadOrStore(normalized, true); !loaded {
				c.queue.Enqueue(queue.Item{URL: normalized, Depth: item.Depth + 1})
			}
		}
	}
}

func (c *Crawph) monitorWorkers() {
	for {
		time.Sleep(100 * time.Millisecond)
		if c.getActive() == 0 && c.queue.IsEmpty() {
			c.queue.Terminate()
			c.logger.Debug("termination signal sent")
			return
		}
	}
}

func (c *Crawph) incActive() {
	c.activeMu.Lock()
	c.activeWorkers++
	c.activeMu.Unlock()
}

func (c *Crawph) decActive() {
	c.activeMu.Lock()
	c.activeWorkers--
	c.activeMu.Unlock()
}

func (c *Crawph) getActive() int {
	c.activeMu.Lock()
	defer c.activeMu.Unlock()
	return c.activeWorkers
}

func extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return parsed.Host
}
