package extractor

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type LinkExtractor struct{}

func (le *LinkExtractor) Name() string {
	return "links"
}

func (le *LinkExtractor) Extract(doc *html.Node, sourceURL string) (*ExtractionResult, error) {
	base, err := url.Parse(sourceURL)
	if err != nil {
		return nil, err
	}
	var links []string
	le.findLinks(doc, base, &links)
	return &ExtractionResult{Links: links}, nil
}

func (le *LinkExtractor) findLinks(node *html.Node, base *url.URL, links *[]string) {
	if node.Type == html.ElementNode && node.Data == "a" {
		for _, attr := range node.Attr {
			if attr.Key != "href" {
				continue
			}
			resolved := le.resolveLink(attr.Val, base)
			if resolved != "" {
				*links = append(*links, resolved)
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		le.findLinks(c, base, links)
	}
}

func (le *LinkExtractor) resolveLink(raw string, base *url.URL) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(parsed)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}
	return resolved.String()
}
