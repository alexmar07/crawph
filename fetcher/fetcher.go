package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/html"
)

type Fetcher struct {
	client    *http.Client
	userAgent string
}

func New(timeout time.Duration, userAgent string, maxRedirects int) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirects {
					return fmt.Errorf("stopped after %d redirects", maxRedirects)
				}
				return nil
			},
		},
		userAgent: userAgent,
	}
}

func (f *Fetcher) Fetch(rawURL string) (*html.Node, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for %s: %w", rawURL, err)
	}
	req.Header.Set("User-Agent", f.userAgent)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", rawURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetching %s: status %d", rawURL, resp.StatusCode)
	}

	// Limit body to 10MB to prevent memory exhaustion on large responses
	const maxBodySize = 10 << 20
	node, err := html.Parse(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML from %s: %w", rawURL, err)
	}

	return node, nil
}
