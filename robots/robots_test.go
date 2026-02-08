package robots

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllowedWhenNoRobotsTxt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	checker := NewChecker("Crawph/1.0", 10*time.Second)
	allowed, err := checker.IsAllowed(server.URL + "/page")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected allowed when no robots.txt")
	}
}

func TestDisallowedPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.Write([]byte("User-agent: *\nDisallow: /private/\n"))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	checker := NewChecker("Crawph/1.0", 10*time.Second)
	allowed, _ := checker.IsAllowed(server.URL + "/private/secret")
	if allowed {
		t.Error("expected disallowed for /private/secret")
	}
	allowed, _ = checker.IsAllowed(server.URL + "/public/page")
	if !allowed {
		t.Error("expected allowed for /public/page")
	}
}

func TestCachesRobotsTxt(t *testing.T) {
	fetchCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			fetchCount++
			w.Write([]byte("User-agent: *\nAllow: /\n"))
			return
		}
	}))
	defer server.Close()
	checker := NewChecker("Crawph/1.0", 10*time.Second)
	checker.IsAllowed(server.URL + "/page1")
	checker.IsAllowed(server.URL + "/page2")
	checker.IsAllowed(server.URL + "/page3")
	if fetchCount != 1 {
		t.Errorf("expected 1 robots.txt fetch, got %d", fetchCount)
	}
}

func TestGetCrawlDelay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.Write([]byte("User-agent: *\nCrawl-delay: 5\n"))
			return
		}
	}))
	defer server.Close()
	checker := NewChecker("Crawph/1.0", 10*time.Second)
	checker.IsAllowed(server.URL + "/page")
	delay := checker.GetCrawlDelay(server.URL)
	if delay != 5*time.Second {
		t.Errorf("expected crawl-delay 5s, got %v", delay)
	}
}

func TestGetCrawlDelayNone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.Write([]byte("User-agent: *\nAllow: /\n"))
			return
		}
	}))
	defer server.Close()
	checker := NewChecker("Crawph/1.0", 10*time.Second)
	checker.IsAllowed(server.URL + "/page")
	delay := checker.GetCrawlDelay(server.URL)
	if delay != 0 {
		t.Errorf("expected crawl-delay 0, got %v", delay)
	}
}

func TestDisabledChecker(t *testing.T) {
	checker := NewDisabledChecker()
	allowed, err := checker.IsAllowed("https://example.com/anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("disabled checker should allow everything")
	}
}
