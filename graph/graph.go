package graph

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
)

type Graph struct {
	Vertices           []*Vertex
	VertexBaseUrlIndex map[string][]*Vertex
	VertexFullUrlIndex map[string]*Vertex
	mu                 sync.RWMutex
}

type Vertex struct {
	BaseUrl string
	FullUrl string
	Edges   []*Vertex
	mu      sync.Mutex
}

func NewGraph() *Graph {
	return &Graph{
		Vertices:           []*Vertex{},
		VertexBaseUrlIndex: map[string][]*Vertex{},
		VertexFullUrlIndex: map[string]*Vertex{},
	}
}

func NewVertex(baseUrl string, fullUrl string) *Vertex {
	return &Vertex{
		BaseUrl: baseUrl,
		FullUrl: fullUrl,
	}
}

// NormalizeURL lowercases scheme/host, removes fragments, removes
// trailing slashes (except root path), and sorts query parameters.
func NormalizeURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid URL %q: missing scheme or host", rawURL)
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""
	if parsed.Path != "/" {
		parsed.Path = strings.TrimRight(parsed.Path, "/")
	}
	if parsed.RawQuery != "" {
		params := parsed.Query()
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sorted := url.Values{}
		for _, k := range keys {
			for _, v := range params[k] {
				sorted.Add(k, v)
			}
		}
		parsed.RawQuery = sorted.Encode()
	}
	return parsed.String(), nil
}

// AddVertex adds a vertex for the given URL, or returns the existing one.
// Uses a single write lock for the entire check-then-add to prevent TOCTOU races.
func (g *Graph) AddVertex(rawURL string) (*Vertex, error) {
	normalized, err := NormalizeURL(rawURL)
	if err != nil {
		return nil, err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if existing, ok := g.VertexFullUrlIndex[normalized]; ok {
		return existing, nil
	}
	parsed, _ := url.Parse(normalized)
	baseUrl := parsed.Scheme + "://" + parsed.Host
	v := NewVertex(baseUrl, normalized)
	g.Vertices = append(g.Vertices, v)
	g.VertexFullUrlIndex[normalized] = v
	g.VertexBaseUrlIndex[baseUrl] = append(g.VertexBaseUrlIndex[baseUrl], v)
	return v, nil
}

func (g *Graph) SearchVertexByFullUrl(fullUrl string) *Vertex {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.VertexFullUrlIndex[fullUrl]
}

func (g *Graph) SearchVertexByBaseUrl(baseUrl string) []*Vertex {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.VertexBaseUrlIndex[baseUrl]
}

func (g *Graph) AddEdge(v1, v2 *Vertex) {
	v1.mu.Lock()
	defer v1.mu.Unlock()
	for _, edge := range v1.Edges {
		if edge.FullUrl == v2.FullUrl {
			return
		}
	}
	v1.Edges = append(v1.Edges, v2)
}
