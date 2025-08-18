package scrapper

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
)

type CollyScraper struct {
	collector *colly.Collector
	timeout   time.Duration
	userAgent string
}

func NewCollyScraper() *CollyScraper {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; MGDS"),
	)

	c.SetRequestTimeout(10 * time.Second)

	return &CollyScraper{
		collector: c,
		timeout:   10 * time.Second,
		userAgent: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; MGDS",
	}
}

func NewCollyScraperWithDebug() *CollyScraper {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; MGDS"),
		colly.Debugger(&debug.LogDebugger{}),
	)

	c.SetRequestTimeout(10 * time.Second)

	return &CollyScraper{
		collector: c,
		timeout:   10 * time.Second,
		userAgent: "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; MGDS",
	}
}

func (cs *CollyScraper) SetUserAgent(userAgent string) {
	cs.userAgent = userAgent
	cs.collector.UserAgent = userAgent
}

func (cs *CollyScraper) SetTimeout(timeout time.Duration) {
	cs.timeout = timeout
	cs.collector.SetRequestTimeout(timeout)
}

func (cs *CollyScraper) Close() error {
	return nil
}

func (cs *CollyScraper) Scrape(ctx context.Context, targetURL string, options *ScrapingOptions) (*ScrapedData, error) {
	if options == nil {
		options = DefaultScrapingOptions()
	}

	result := &ScrapedData{
		URL:        targetURL,
		ScrapedAt:  time.Now(),
		Links:      []Link{},
		MetaTags:   make(map[string]string),
		Headers:    make(map[string]string),
	}

	c := cs.collector.Clone()

	cs.configureCollector(c, options)

	c.OnResponse(func(r *colly.Response) {
		result.StatusCode = r.StatusCode
		for key, values := range *r.Headers {
			if len(values) > 0 {
				result.Headers[key] = values[0]
			}
		}

		if options.ExtractHTML {
			result.HTMLBody = string(r.Body)
		}
	})

	c.OnHTML("html", func(e *colly.HTMLElement) {
		if options.ExtractText {
			result.Text = cs.extractText(e)
		}

		if options.ExtractMeta {
			result.Title = e.ChildText("head title")
			cs.extractMetaTags(e, result)
		}

		if options.ExtractLinks {
			cs.extractLinks(e, result)
		}

	})

	c.OnError(func(r *colly.Response, err error) {
		result.StatusCode = r.StatusCode
	})

	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- c.Visit(targetURL)
	}()

	select {
	case err := <-done:
		if err != nil {
			return result, fmt.Errorf("failed to scrape %s: %v", targetURL, err)
		}
		return result, nil
	case <-ctx.Done():
		return result, fmt.Errorf("scraping timeout for %s", targetURL)
	}
}

func (cs *CollyScraper) ScrapeMultiple(ctx context.Context, urls []string, options *ScrapingOptions) ([]*ScrapedData, error) {
	results := make([]*ScrapedData, 0, len(urls))

	for _, url := range urls {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
			data, err := cs.Scrape(ctx, url, options)
			if err != nil {
				data = &ScrapedData{
					URL:        url,
					StatusCode: 0,
					ScrapedAt:  time.Now(),
				}
			}
			results = append(results, data)

			if options != nil && options.RateLimitDelay > 0 {
				time.Sleep(options.RateLimitDelay)
			}
		}
	}

	return results, nil
}

func (cs *CollyScraper) configureCollector(c *colly.Collector, options *ScrapingOptions) {
	c.SetRequestTimeout(options.Timeout)
	c.UserAgent = options.UserAgent

	if len(options.AllowedDomains) > 0 {
		c.AllowedDomains = options.AllowedDomains
	}

	if len(options.DisallowedDomains) > 0 {
		c.DisallowedDomains = options.DisallowedDomains
	}

	if options.MaxDepth > 0 {
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			Parallelism: 1,
			Delay:       options.RateLimitDelay,
		})
	}

	if !options.FollowRedirects {
		c.OnResponse(func(r *colly.Response) {
			if r.StatusCode >= 300 && r.StatusCode < 400 {
				r.Request.Abort()
			}
		})
	}
}

func (cs *CollyScraper) extractText(e *colly.HTMLElement) string {
	var texts []string

	e.ForEach("p, h1, h2, h3, h4, h5, h6, div, span, article, section", func(i int, elem *colly.HTMLElement) {
		text := strings.TrimSpace(elem.Text)
		if text != "" && len(text) > 3 {
			texts = append(texts, text)
		}
	})

	return strings.Join(texts, " ")
}

func (cs *CollyScraper) extractMetaTags(e *colly.HTMLElement, result *ScrapedData) {
	e.ForEach("meta", func(i int, elem *colly.HTMLElement) {
		name := elem.Attr("name")
		property := elem.Attr("property")
		content := elem.Attr("content")

		if name != "" && content != "" {
			result.MetaTags[name] = content
		}
		if property != "" && content != "" {
			result.MetaTags[property] = content
		}
	})
}

func (cs *CollyScraper) extractLinks(e *colly.HTMLElement, result *ScrapedData) {
	baseURL, _ := url.Parse(result.URL)

	e.ForEach("a[href]", func(i int, elem *colly.HTMLElement) {
		href := elem.Attr("href")
		if href == "" {
			return
		}

		absoluteURL := cs.resolveURL(baseURL, href)

		link := Link{
			URL:          absoluteURL,
			Text:         strings.TrimSpace(elem.Text),
			Rel:          elem.Attr("rel"),
			Target:       elem.Attr("target"),
			Download:     elem.Attr("download"),
			ResourceType: cs.detectResourceType(absoluteURL),
			IsExternal:   cs.isExternalDomain(result.URL, absoluteURL),
		}

		result.Links = append(result.Links, link)
	})
}





func (cs *CollyScraper) resolveURL(base *url.URL, reference string) string {
	if base == nil {
		return reference
	}

	u, err := url.Parse(reference)
	if err != nil {
		return reference
	}

	return base.ResolveReference(u).String()
}

func (cs *CollyScraper) detectResourceType(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	path := strings.ToLower(u.Path)
	if idx := strings.LastIndex(path, "."); idx != -1 {
		ext := path[idx+1:]
		switch ext {
		case "js":
			return "js"
		case "css":
			return "css"
		case "pdf":
			return "pdf"
		case "doc", "docx":
			return "document"
		case "xls", "xlsx":
			return "spreadsheet"
		case "ppt", "pptx":
			return "presentation"
		case "zip", "rar", "tar", "gz", "7z":
			return "archive"
		case "jpg", "jpeg", "png", "gif", "svg", "webp":
			return "image"
		case "json", "xml":
			return "data"
		case "txt", "log":
			return "text"
		}
	}
	return ""
}

func (cs *CollyScraper) isExternalDomain(baseURL, linkURL string) bool {
	base, err := url.Parse(baseURL)
	if err != nil {
		return false
	}

	link, err := url.Parse(linkURL)
	if err != nil {
		return false
	}

	// Si le lien n'a pas de host (relatif), ce n'est pas externe
	if link.Host == "" {
		return false
	}

	// Si le scheme n'est pas http/https, considérer comme non-externe pour l'OSINT
	if link.Scheme != "http" && link.Scheme != "https" {
		return false
	}

	// Comparer les domaines
	return base.Host != link.Host
}
