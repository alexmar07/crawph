package robots

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
)

type RobotsChecker interface {
	IsAllowed(rawURL string) (bool, error)
	GetCrawlDelay(rawURL string) time.Duration
}

type Checker struct {
	userAgent string
	client    *http.Client
	cache     map[string]*robotstxt.RobotsData
	mu        sync.Mutex
}

func NewChecker(userAgent string, timeout time.Duration) *Checker {
	return &Checker{
		userAgent: userAgent,
		client:    &http.Client{Timeout: timeout},
		cache:     make(map[string]*robotstxt.RobotsData),
	}
}

func (c *Checker) IsAllowed(rawURL string) (bool, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false, fmt.Errorf("parsing URL %s: %w", rawURL, err)
	}
	origin := parsed.Scheme + "://" + parsed.Host
	robots, err := c.getRobots(origin)
	if err != nil {
		return true, nil
	}
	group := robots.FindGroup(c.userAgent)
	return group.Test(parsed.Path), nil
}

func (c *Checker) GetCrawlDelay(rawURL string) time.Duration {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return 0
	}
	origin := parsed.Scheme + "://" + parsed.Host
	c.mu.Lock()
	robots, ok := c.cache[origin]
	c.mu.Unlock()
	if !ok {
		return 0
	}
	group := robots.FindGroup(c.userAgent)
	return group.CrawlDelay
}

func (c *Checker) getRobots(origin string) (*robotstxt.RobotsData, error) {
	c.mu.Lock()
	if robots, ok := c.cache[origin]; ok {
		c.mu.Unlock()
		if robots == nil {
			// Sentinel: fetch in progress or failed
			return &robotstxt.RobotsData{}, nil
		}
		return robots, nil
	}
	// Mark as in-progress with nil sentinel
	c.cache[origin] = nil
	c.mu.Unlock()

	robotsURL := origin + "/robots.txt"
	resp, err := c.client.Get(robotsURL)
	if err != nil {
		c.mu.Lock()
		delete(c.cache, origin)
		c.mu.Unlock()
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var robots *robotstxt.RobotsData
	if resp.StatusCode != http.StatusOK {
		robots = &robotstxt.RobotsData{}
	} else {
		robots, err = robotstxt.FromResponse(resp)
		if err != nil {
			c.mu.Lock()
			delete(c.cache, origin)
			c.mu.Unlock()
			return nil, fmt.Errorf("parsing robots.txt from %s: %w", origin, err)
		}
	}

	c.mu.Lock()
	c.cache[origin] = robots
	c.mu.Unlock()

	return robots, nil
}

// DisabledChecker always allows crawling.
type DisabledChecker struct{}

func NewDisabledChecker() *DisabledChecker {
	return &DisabledChecker{}
}

func (d *DisabledChecker) IsAllowed(_ string) (bool, error) {
	return true, nil
}

func (d *DisabledChecker) GetCrawlDelay(_ string) time.Duration {
	return 0
}
