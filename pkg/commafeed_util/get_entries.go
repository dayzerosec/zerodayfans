package commafeed_util

import (
	"context"
	"fmt"
	"github.com/dayzerosec/zerodayfans/pkg/commafeed"
	"github.com/dayzerosec/zerodayfans/pkg/config"
	"log"
	"regexp"
	"strings"
	"time"
)

type GetEntriesOptions struct {
	// Feed contains the Config for the specific feed we want to retrieve
	Feed config.FeedConfig

	// NewerThan reflects the `insertedDate` of the entry and not the published date so on a
	// So any benefit of trimming won't be seen until the instance has been running long enough
	NewerThan time.Time

	// MaxEntries is the maximum number of entries for commafeed to return, this should be set
	// to the number of entries that will be in the final feed so you don't accidentally trim any
	MaxEntries int
}

// GetRawFeedEntries will return the Commafeed source entries for a given feed (or category) without apply any filters
func GetRawFeedEntries(ctx context.Context, client *commafeed.ClientWithResponses, opts GetEntriesOptions) (*commafeed.Entries, error) {
	newerThan64 := opts.NewerThan.UnixMilli()
	maxEntries32 := int32(opts.MaxEntries)

	if maxEntries32 == 0 {
		log.Println("[WARN] MaxEntries is 0, this will not return any entries")
	}

	var entries *commafeed.Entries
	switch opts.Feed.Type {
	case config.FeedTypeCategory:
		res, err := client.GetCategoryEntriesWithResponse(ctx, &commafeed.GetCategoryEntriesParams{
			Id:        opts.Feed.ID,
			ReadType:  "all",
			NewerThan: &newerThan64,
			Limit:     &maxEntries32,
		})
		if err != nil {
			return nil, err
		}
		if res.StatusCode() != 200 {
			return nil, fmt.Errorf("unexpected status code(%d)", res.StatusCode())
		}
		entries = res.JSONDefault
		trackCategory(ctx, client, opts.Feed.ID)
	case config.FeedTypeFeed:
		res, err := client.GetFeedEntriesWithResponse(ctx, &commafeed.GetFeedEntriesParams{
			Id:        opts.Feed.ID,
			ReadType:  "all",
			NewerThan: &newerThan64,
			Limit:     &maxEntries32,
		})
		if err != nil {
			return nil, err
		}
		if res.StatusCode() != 200 {
			return nil, fmt.Errorf("unexpected status code(%d)", res.StatusCode())
		}
		entries = res.JSONDefault
		if len(entries.Entries) > 0 {
			e := entries.Entries[0]
			trackFeeds(e.FeedName, e.FeedLink, e.FeedUrl)
		}
	}

	return entries, nil
}

// FilterEntries will take a given entries list and return an Entry array that has been filtered based on the feed config
func FilterEntries(feed config.FeedConfig, entries *commafeed.Entries) ([]commafeed.Entry, error) {
	var filtered []commafeed.Entry

	for _, entry := range entries.Entries {
		if dailyLimitReached(entry, filtered, feed.DailyLimit) {
			continue
		}
		doUrlRewrite(&entry, feed)
		cleanFeedTitle(&entry)

		if len(feed.Filters) > 0 {
			result := true
			for _, filter := range feed.Filters {
				result = result && doFilter(entry, filter)
			}
			if result {
				filtered = append(filtered, entry)
			}
		} else {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// GetEntries will return a commafeed.Entries object with Entries[] that have been filtered and processed
func GetEntries(ctx context.Context, client *commafeed.ClientWithResponses, opts GetEntriesOptions) (*commafeed.Entries, error) {
	entries, err := GetRawFeedEntries(ctx, client, opts)
	if err != nil {
		return nil, err
	}

	filtered, err := FilterEntries(opts.Feed, entries)
	if err != nil {
		return nil, err
	}

	entries.Entries = filtered
	return entries, nil
}

// dailyLimitReached will determine if the current entry should be skipped due to the daily limit being reached
// returns True if the limit has already been reached, false if there is no limit or the limit hasn't been reached
func dailyLimitReached(newEntry commafeed.Entry, entries []commafeed.Entry, limit int) bool {
	if limit == 0 {
		return false
	}
	if len(entries) < limit {
		return false
	}

	lastEntries := entries[len(entries)-limit:]

	pastTime, _ := PrimitiveToTime(TimePrimitive(lastEntries[0].Date))
	newTime, _ := PrimitiveToTime(TimePrimitive(newEntry.Date))

	// Since processing in descending order, pastTime is technically the newer time
	delta := pastTime.Sub(newTime)
	return delta < 24*time.Hour
}

// doUrlRewrite will apply the URL rewrite transformations from the feed config to the given entry
func doUrlRewrite(entry *commafeed.Entry, feed config.FeedConfig) {
	if len(feed.UrlRewrite) != 2 {
		return
	}

	match, err := regexp.Compile(feed.UrlRewrite[0])
	if err != nil {
		log.Printf("Error compiling regex: '%s' -> %v", feed.UrlRewrite[0], err)
		return
	}

	entry.FeedUrl = match.ReplaceAllString(entry.FeedUrl, feed.UrlRewrite[1])
	entry.Url = match.ReplaceAllString(entry.Url, feed.UrlRewrite[1])
}

// cleanFeedTitle will remove any (...) content from the end of a title, usually indicative of info I've added in commafeed
// and not intended to be part of the true title
func cleanFeedTitle(entry *commafeed.Entry) {
	title := strings.TrimSpace(entry.Title)
	if strings.HasSuffix(title, ")") {
		idx := strings.LastIndex(title, "(")
		lenInsideBrackets := len(title) - idx
		if lenInsideBrackets > 1 && lenInsideBrackets < 10 {
			entry.Title = strings.TrimSpace(title[:idx])
		}
	}

}
