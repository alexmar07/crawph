package crawph

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/alexmar07/crawler-go/extractor"
	"github.com/alexmar07/crawler-go/robots"
)

func TestCrawlEndToEnd(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body>
			<a href="/page1">Page 1</a>
			<a href="/page2">Page 2</a>
		</body></html>`))
	})
	mux.HandleFunc("/page1", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body>
			<a href="/page3">Page 3</a>
		</body></html>`))
	})
	mux.HandleFunc("/page2", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body><p>Dead end</p></body></html>`))
	})
	mux.HandleFunc("/page3", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body>
			<a href="/">Back to home</a>
		</body></html>`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	crawler := New(Options{
		MaxWorkers:    2,
		MaxDepth:      5,
		Timeout:       10 * time.Second,
		UserAgent:     "TestBot/1.0",
		DefaultRPS:    100.0,
		RobotsChecker: robots.NewDisabledChecker(),
		Extractors:    []extractor.Extractor{&extractor.LinkExtractor{}},
		Logger:        logger,
	})

	crawler.Start([]string{server.URL + "/"})

	g := crawler.Graph()

	if len(g.Vertices) != 4 {
		urls := make([]string, len(g.Vertices))
		for i, v := range g.Vertices {
			urls[i] = v.FullUrl
		}
		t.Errorf("expected 4 vertices, got %d: %v", len(g.Vertices), urls)
	}

	root := g.SearchVertexByFullUrl(server.URL + "/")
	if root == nil {
		root = g.SearchVertexByFullUrl(server.URL)
	}
	if root == nil {
		t.Fatal("root vertex not found")
	}
	if len(root.Edges) != 2 {
		t.Errorf("expected 2 edges from root, got %d", len(root.Edges))
	}
}

func TestCrawlRespectsMaxDepth(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body><a href="/level1">L1</a></body></html>`))
	})
	mux.HandleFunc("/level1", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body><a href="/level2">L2</a></body></html>`))
	})
	mux.HandleFunc("/level2", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body><a href="/level3">L3</a></body></html>`))
	})
	mux.HandleFunc("/level3", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body><p>Too deep</p></body></html>`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	crawler := New(Options{
		MaxWorkers:    1,
		MaxDepth:      1,
		Timeout:       10 * time.Second,
		UserAgent:     "TestBot/1.0",
		DefaultRPS:    100.0,
		RobotsChecker: robots.NewDisabledChecker(),
		Extractors:    []extractor.Extractor{&extractor.LinkExtractor{}},
		Logger:        logger,
	})

	crawler.Start([]string{server.URL + "/"})

	g := crawler.Graph()

	hasLevel3 := g.SearchVertexByFullUrl(server.URL + "/level3")
	if hasLevel3 != nil {
		t.Error("level3 should not exist — max depth is 1")
	}
}
