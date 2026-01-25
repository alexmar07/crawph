package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func TestFetchReturnsHTMLNode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body><a href=\"/link\">Link</a></body></html>"))
	}))
	defer server.Close()
	f := New(30*time.Second, "Crawph/1.0", 10)
	node, err := f.Fetch(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil html.Node")
	}
	if node.Type != html.DocumentNode {
		t.Errorf("expected DocumentNode, got %v", node.Type)
	}
}

func TestFetchSetsUserAgent(t *testing.T) {
	var receivedUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.Write([]byte("<html></html>"))
	}))
	defer server.Close()
	f := New(30*time.Second, "TestBot/1.0", 10)
	f.Fetch(server.URL)
	if receivedUA != "TestBot/1.0" {
		t.Errorf("expected User-Agent TestBot/1.0, got %s", receivedUA)
	}
}

func TestFetchTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte("<html></html>"))
	}))
	defer server.Close()
	f := New(100*time.Millisecond, "Crawph/1.0", 10)
	_, err := f.Fetch(server.URL)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestFetchNon2xxStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	f := New(30*time.Second, "Crawph/1.0", 10)
	_, err := f.Fetch(server.URL)
	if err == nil {
		t.Error("expected error for 404 status")
	}
}

func TestFetchAccepts200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html></html>"))
	}))
	defer server.Close()
	f := New(30*time.Second, "Crawph/1.0", 10)
	_, err := f.Fetch(server.URL)
	if err != nil {
		t.Errorf("expected no error for 200, got %v", err)
	}
}

func TestFetchRedirectLimit(t *testing.T) {
	redirectCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectCount++
		if redirectCount <= 20 {
			http.Redirect(w, r, "/redirect", http.StatusFound)
			return
		}
		w.Write([]byte("<html></html>"))
	}))
	defer server.Close()
	f := New(30*time.Second, "Crawph/1.0", 3)
	_, err := f.Fetch(server.URL)
	if err == nil {
		t.Error("expected error when redirect limit exceeded")
	}
}

func TestFetchInvalidURL(t *testing.T) {
	f := New(30*time.Second, "Crawph/1.0", 10)
	_, err := f.Fetch("://invalid")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}
