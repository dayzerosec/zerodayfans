package enrichment

import (
	"errors"
	"fmt"
	"github.com/dayzerosec/zerodayfans/pkg/cache"
	"github.com/dayzerosec/zerodayfans/pkg/commafeed"
	"github.com/dayzerosec/zerodayfans/pkg/config"
	"log"
	"net/url"
	"strings"
	"time"
)

var c cache.ObjectCache
var ErrCantEnrich = errors.New("can't enrich this url")

type EnrichedData struct {
	Publisher SiteInfo
	Author    AuthorInfo
	Page      PageInfo
	Feed      FeedInfo
}

type SiteInfo struct {
	Name  string
	Url   string
	Image string
}

type AuthorInfo struct {
	Name string
	Url  string
}

type PageInfo struct {
	Title       string
	Description string
	Image       string
	Url         string
	Content     string
}

type FeedInfo struct {
	Title       string
	Source      string
	Description string
	Date        time.Time
}

type EnrichFn func(string) (*EnrichedData, error)

var enrichers = map[string]EnrichFn{}

func RegisterEnricher(target string, fn EnrichFn) {
	enrichers[target] = fn
}

// FindCustomEnrichment will either return the enrichment function, or nil if there isn't a custom function for this target
func FindCustomEnrichment(target string) (EnrichFn, error) {
	target = strings.ToLower(target)
	u, err := url.Parse(target)
	if u == nil || err != nil {
		return nil, err
	}

	if fn, found := enrichers[u.Host]; found {
		return fn, nil
	}

	return nil, nil
}

func initCache() error {
	if config.Cfg.EnrichmentCache.File == "" {
		panic("enrichment cache file must be specified")
	} else {
		cacheMaxAge := time.Hour * 24 * 14
		var err error
		c, err = cache.NewJsonCache(config.Cfg.EnrichmentCache)
		if err != nil {
			return err
		}
		if err = c.Clean(cacheMaxAge); err != nil {
			log.Println("Error cleaning cache:", err)
		}
	}
	return nil
}

func Enrich(entry commafeed.Entry) (out *EnrichedData, err error) {
	if c == nil {
		if err = initCache(); err != nil {
			return nil, err
		}
	}

	target := entry.Url
	out = &EnrichedData{}
	if found := c.Get(target, out); found {
		return out, nil
	}

	out, err = getBaseEnrichment(target)
	if err != nil {
		return nil, err
	}
	doFeedEnrichment(out, entry)

	if err == nil {
		_ = c.Set(target, out)
	}

	return
}

func getBaseEnrichment(target string) (out *EnrichedData, err error) {
	if fn, _ := FindCustomEnrichment(target); fn != nil {
		if out, err = fn(target); err == nil {
			return
		}
	}

	// Custom option failed or didn't exist so fallback to default
	out, err = DefaultEnricher(target)
	return
}

func doFeedEnrichment(enriched *EnrichedData, e commafeed.Entry) {
	enriched.Feed.Date = time.UnixMilli(int64(e.Date))
	enriched.Feed.Title = e.Title
	enriched.Feed.Source = e.FeedUrl

	// Title and URL from feed take precedence over anything parsed
	enriched.Page.Title = e.Title
	enriched.Page.Url = e.Url

	// Populate fields not already populated with feed data so we don't have a completely empty enrichment
	if enriched.Publisher.Name == "" {
		enriched.Publisher.Name = e.FeedName
	}
	if enriched.Publisher.Url == "" {
		if e.FeedLink != e.FeedUrl {
			enriched.Publisher.Url = e.FeedLink
		} else {
			u, _ := url.Parse(e.Url)
			enriched.Publisher.Url = fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		}
	}

	if enriched.Publisher.Image == "" {
		enriched.Publisher.Image = config.Cfg.Output.DefaultFavicon
	}

	if enriched.Author.Name == "" {
		if e.Author != nil && *e.Author != "" {
			enriched.Author.Name = *e.Author
		}
	}

	// Only Take content from feed if we didn't already manage to scrape something
	if enriched.Page.Content == "" && e.Content != "" {
		enriched.Page.Content = e.Content
	}

	if enriched.Author.Url == "" {
		// The feed cannot provide an author URL
	}
}

func GetFeedTopper() *EnrichedData {
	if c == nil {
		if err := initCache(); err != nil {
			log.Printf("Error initializing cache: %v", err)
			return nil
		}
	}

	out := &EnrichedData{}
	if !c.Get("__newest", out) {
		return nil
	}
	return out
}
func SetFeedTopper(data *EnrichedData) error {
	return c.Set("__newest", data)
}
