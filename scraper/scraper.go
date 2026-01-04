package scraper

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type Scraper struct {
	rawUrl          string
	url             *url.URL
	discoveredLinks []string
}

func NewScraper(rawUrl string) *Scraper {

	url, _ := url.Parse(rawUrl)

	return &Scraper{
		rawUrl: rawUrl,
		url:    url,
	}
}

func (s *Scraper) StartDiscovered() (int, error) {

	response, err := http.Get(s.rawUrl)

	if err != nil {
		return 0, err
	}

	if response.StatusCode != 200 {
		return 0, errors.New("Il codice di risposta è diverso da 200")
	}

	defer response.Body.Close()

	rootNode, _ := html.Parse(response.Body)

	s.discoveredLinks = s.findLinks(rootNode)

	return len(s.discoveredLinks), nil

}

func (s Scraper) GetDiscoveredLinks() []string {
	return s.discoveredLinks
}

func (s Scraper) CountDiscoveredLinks() int {
	return len(s.discoveredLinks)
}

func (s *Scraper) findLinks(node *html.Node) (links []string) {

	if node.Type == html.ElementNode && node.Data == "a" {
		for _, attr := range node.Attr {
			if attr.Key != "href" {
				continue
			}

			if strings.HasPrefix(attr.Val, "https://") || strings.HasPrefix(attr.Val, "http://") || strings.HasPrefix(attr.Val, "/") {
				links = append(links, s.completeLink(attr.Val))
			}

		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		links = append(links, s.findLinks(c)...)
	}

	return links
}

func (s Scraper) completeLink(link string) string {

	if !strings.HasPrefix(link, "/") {
		return link
	}

	return s.url.Scheme + "://" + s.url.Host + link
}
