package commafeed_util

import (
	"context"
	"fmt"
	"github.com/dayzerosec/zerodayfans/pkg/commafeed"
	"github.com/dayzerosec/zerodayfans/pkg/config"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type trackedFeedEntry struct {
	Name   string
	Link   string
	RssUrl string
}

var TrackedFeeds []trackedFeedEntry

var rootCategory *commafeed.Category

func Login(ctx context.Context, client *commafeed.ClientWithResponses) error {
	/*
		res, err := client.LoginWithResponse(ctx, commafeed.LoginJSONRequestBody{
			Name:     config.Cfg.Commafeed.Username,
			Password: config.Cfg.Commafeed.Password,
		})
		if err != nil {
			return err
		}
		if res.StatusCode() != http.StatusOK {
			return fmt.Errorf("unexpected status code (%d)", res.StatusCode())
		}
	*/
	return nil
}

func getRootCategory(ctx context.Context, client *commafeed.ClientWithResponses) error {
	if config.Cfg.Commafeed.Username == "" || config.Cfg.Commafeed.Password == "" {
		return fmt.Errorf("commafeed credentials are required for tracking categories")
	}
	res, err := client.GetRootCategoryWithResponse(ctx)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code (%d)", res.StatusCode())
	}
	rootCategory = res.JSONDefault
	return nil
}

func trackCategory(ctx context.Context, client *commafeed.ClientWithResponses, id string) {
	if config.Cfg.Output.FeedlistFile == "" {
		return
	}
	if rootCategory == nil {
		if err := getRootCategory(ctx, client); err != nil {
			log.Printf("Error getting root category: %v", err)
			// Can't track without root category
			return
		}
	}

	cat := findCategory(id)
	if cat == nil {
		log.Printf("Category not found: %s", id)
		return
	}

	for _, feed := range cat.Feeds {
		trackFeeds(feed.Name, feed.FeedLink, feed.FeedUrl)
	}

}

func findCategory(id string) *commafeed.Category {
	if rootCategory == nil {
		return nil
	}
	return findCategoryInTree(id, rootCategory)
}

func findCategoryInTree(id string, category *commafeed.Category) *commafeed.Category {
	if category.Id == id {
		return category
	}
	for _, child := range category.Children {
		if found := findCategoryInTree(id, &child); found != nil {
			return found
		}
	}
	return nil
}

func trackFeeds(feedName, siteUrl, feedUrl string) {
	if config.Cfg.Output.FeedlistFile == "" {
		return
	}
	if siteUrl == "" || strings.HasPrefix(siteUrl, "/") {
		if strings.HasPrefix(feedUrl, "http") || strings.HasPrefix(feedUrl, "//") {
			u, _ := url.Parse(feedUrl)
			// Hopefully this isn't a proxy URL like a feedburner link, but screw it, this is already a hack
			// for sites that break the RSS spec by using relative URLs
			siteUrl = fmt.Sprintf("https://%s", u.Hostname())
		}

	}
	TrackedFeeds = append(TrackedFeeds, trackedFeedEntry{
		Name:   feedName,
		Link:   siteUrl,
		RssUrl: feedUrl,
	})
}
