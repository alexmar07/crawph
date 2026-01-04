package graph

import (
	u "net/url"
	"sync"
)

type Graph struct {
	Verticies          []*Vertex
	VertexBaseUrlIndex map[string][]*Vertex
	VertexFullUrlIndex map[string]*Vertex
	mu                 sync.RWMutex
	vMu                sync.RWMutex
	vBaseUrlMu         sync.RWMutex
	vFullUrlMu         sync.RWMutex
}

type Vertex struct {
	BaseUrl string
	FullUrl string
	Edges   []*Vertex
	mu      sync.RWMutex
}

func NewGraph() *Graph {
	return &Graph{
		Verticies:          []*Vertex{},
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

func (g *Graph) AddVertex(url string) *Vertex {

	existVertex := g.SearchVertexByFullUrl(url)

	if existVertex != nil {
		return existVertex
	}

	urlParsed, _ := u.Parse(url)

	// @TODO: Controllo errori

	baseUrl := urlParsed.Scheme + "://" + urlParsed.Host

	v := NewVertex(baseUrl, url)

	g.vMu.Lock()
	g.Verticies = append(g.Verticies, v)
	g.vMu.Unlock()

	g.addUrlIndex(v, baseUrl, url)

	return v
}

func (g *Graph) SearchVertexByFullUrl(fullUrl string) *Vertex {

	g.vFullUrlMu.RLock()

	defer g.vFullUrlMu.RUnlock()

	return g.VertexFullUrlIndex[fullUrl]
}

func (g *Graph) SearchVertexByBaseUrl(baseUrl string) []*Vertex {

	g.vBaseUrlMu.RLock()

	defer g.vBaseUrlMu.RUnlock()

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

func (g *Graph) addUrlIndex(v *Vertex, baseUrl, url string) {

	g.vFullUrlMu.Lock()
	g.VertexFullUrlIndex[url] = v
	g.vFullUrlMu.Unlock()

	g.vBaseUrlMu.Lock()
	g.VertexBaseUrlIndex[baseUrl] = append(g.VertexBaseUrlIndex[baseUrl], v)
	g.vBaseUrlMu.Unlock()
}
