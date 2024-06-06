package feedgen

import (
	"github.com/dayzerosec/zerodayfans/pkg/config"
	"github.com/dayzerosec/zerodayfans/pkg/enrichment"
	"github.com/gorilla/feeds"
	"time"
)

func feedLink(u string) *feeds.Link {
	return &feeds.Link{
		Href: u,
	}
}

func createEmptyFeed() *feeds.Feed {
	feed := &feeds.Feed{
		Title:       config.Cfg.Output.Title,
		Link:        feedLink(config.Cfg.Output.Url),
		Description: config.Cfg.Output.Description,
		Author:      &feeds.Author{Name: config.Cfg.Output.Author},
		Updated:     time.Time{},
		Created:     time.Time{},
		Id:          config.Cfg.Output.Url,
		Subtitle:    config.Cfg.Output.Subtitle,
	}
	return feed
}

func generateFeedItem(data *enrichment.EnrichedData) *feeds.Item {
	item := &feeds.Item{
		Title:       data.Page.Title,
		Link:        feedLink(data.Page.Url),
		Source:      feedLink(data.Feed.Source),
		Author:      &feeds.Author{Name: data.Author.Name},
		Description: data.Page.Description,
		Id:          data.Page.Url,
		Updated:     data.Feed.Date,
		Created:     data.Feed.Date,
		Enclosure:   nil,
		Content:     data.Page.Content,
	}
	return item
}
