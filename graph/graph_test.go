package graph

import (
	"sync"
	"testing"
)

func TestAddVertexAndSearch(t *testing.T) {
	g := NewGraph()
	v, err := g.AddVertex("https://example.com/page")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.FullUrl != "https://example.com/page" {
		t.Errorf("expected https://example.com/page, got %s", v.FullUrl)
	}
	if v.BaseUrl != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", v.BaseUrl)
	}
}

func TestAddVertexDeduplication(t *testing.T) {
	g := NewGraph()
	v1, _ := g.AddVertex("https://example.com/page")
	v2, _ := g.AddVertex("https://example.com/page")
	if v1 != v2 {
		t.Error("expected same vertex for duplicate URL")
	}
	if len(g.Vertices) != 1 {
		t.Errorf("expected 1 vertex, got %d", len(g.Vertices))
	}
}

func TestAddVertexNormalization(t *testing.T) {
	g := NewGraph()
	v1, _ := g.AddVertex("https://Example.COM/page/")
	v2, _ := g.AddVertex("https://example.com/page")
	if v1 != v2 {
		t.Error("normalized URLs should produce the same vertex")
	}
}

func TestAddVertexFragmentRemoval(t *testing.T) {
	g := NewGraph()
	v1, _ := g.AddVertex("https://example.com/page#section")
	v2, _ := g.AddVertex("https://example.com/page")
	if v1 != v2 {
		t.Error("URLs differing only by fragment should be the same vertex")
	}
}

func TestAddVertexInvalidURL(t *testing.T) {
	g := NewGraph()
	_, err := g.AddVertex("://invalid")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestAddEdge(t *testing.T) {
	g := NewGraph()
	v1, _ := g.AddVertex("https://example.com/a")
	v2, _ := g.AddVertex("https://example.com/b")
	g.AddEdge(v1, v2)
	if len(v1.Edges) != 1 || v1.Edges[0] != v2 {
		t.Error("expected edge from v1 to v2")
	}
}

func TestAddEdgeDeduplication(t *testing.T) {
	g := NewGraph()
	v1, _ := g.AddVertex("https://example.com/a")
	v2, _ := g.AddVertex("https://example.com/b")
	g.AddEdge(v1, v2)
	g.AddEdge(v1, v2)
	if len(v1.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(v1.Edges))
	}
}

func TestSearchVertexByBaseUrl(t *testing.T) {
	g := NewGraph()
	g.AddVertex("https://example.com/a")
	g.AddVertex("https://example.com/b")
	g.AddVertex("https://other.com/c")
	results := g.SearchVertexByBaseUrl("https://example.com")
	if len(results) != 2 {
		t.Errorf("expected 2 vertices for example.com, got %d", len(results))
	}
}

func TestConcurrentAddVertex(t *testing.T) {
	g := NewGraph()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.AddVertex("https://example.com/same")
		}()
	}
	wg.Wait()
	if len(g.Vertices) != 1 {
		t.Errorf("expected 1 vertex after concurrent adds, got %d", len(g.Vertices))
	}
}
