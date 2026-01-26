package extractor

import (
	"golang.org/x/net/html"
)

type ExtractionResult struct {
	Links []string
	Data  map[string]any
}

type Extractor interface {
	Name() string
	Extract(doc *html.Node, sourceURL string) (*ExtractionResult, error)
}
