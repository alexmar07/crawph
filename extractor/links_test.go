package extractor

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func parseHTML(t *testing.T, raw string) *html.Node {
	t.Helper()
	node, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}
	return node
}

func TestLinkExtractorFindsAbsoluteLinks(t *testing.T) {
	doc := parseHTML(t, `<html><body>
		<a href="https://example.com/page1">P1</a>
		<a href="https://example.com/page2">P2</a>
	</body></html>`)
	le := &LinkExtractor{}
	result, err := le.Extract(doc, "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Links) != 2 {
		t.Errorf("expected 2 links, got %d", len(result.Links))
	}
}

func TestLinkExtractorResolvesRelativeLinks(t *testing.T) {
	doc := parseHTML(t, `<html><body><a href="/about">About</a></body></html>`)
	le := &LinkExtractor{}
	result, err := le.Extract(doc, "https://example.com/page")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Links) != 1 {
		t.Errorf("expected 1 link, got %d", len(result.Links))
	}
	if result.Links[0] != "https://example.com/about" {
		t.Errorf("expected https://example.com/about, got %s", result.Links[0])
	}
}

func TestLinkExtractorIgnoresNonHTTPLinks(t *testing.T) {
	doc := parseHTML(t, `<html><body>
		<a href="mailto:test@test.com">Email</a>
		<a href="javascript:void(0)">JS</a>
		<a href="tel:+123">Phone</a>
		<a href="https://example.com">Valid</a>
	</body></html>`)
	le := &LinkExtractor{}
	result, err := le.Extract(doc, "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Links) != 1 {
		t.Errorf("expected 1 link, got %d: %v", len(result.Links), result.Links)
	}
}

func TestLinkExtractorNoLinks(t *testing.T) {
	doc := parseHTML(t, `<html><body><p>No links here</p></body></html>`)
	le := &LinkExtractor{}
	result, err := le.Extract(doc, "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Links) != 0 {
		t.Errorf("expected 0 links, got %d", len(result.Links))
	}
}

func TestLinkExtractorName(t *testing.T) {
	le := &LinkExtractor{}
	if le.Name() != "links" {
		t.Errorf("expected name 'links', got %s", le.Name())
	}
}

func TestLinkExtractorNestedLinks(t *testing.T) {
	doc := parseHTML(t, `<html><body>
		<div><ul><li><a href="https://example.com/deep">Deep</a></li></ul></div>
	</body></html>`)
	le := &LinkExtractor{}
	result, err := le.Extract(doc, "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Links) != 1 {
		t.Errorf("expected 1 link, got %d", len(result.Links))
	}
}
