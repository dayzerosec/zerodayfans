package feedgen

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dayzerosec/zerodayfans/pkg/commafeed"
	"github.com/dayzerosec/zerodayfans/pkg/commafeed_util"
	"github.com/dayzerosec/zerodayfans/pkg/config"
	"github.com/dayzerosec/zerodayfans/pkg/enrichment"
	"github.com/gorilla/feeds"
	"log"
	"os"
	"path/filepath"
	"time"
)

func getFeedEntries() []commafeed.Entry {
	ctx := context.Background()
	c, _ := config.Cfg.Commafeed.Client()

	_ = commafeed_util.Login(ctx, c)

	opts := commafeed_util.GetEntriesOptions{
		Feed:       config.FeedConfig{},
		NewerThan:  time.Time{},
		MaxEntries: config.Cfg.MaxEntries,
	}

	var mergedEntries []commafeed.Entry
	for key, feed := range config.Cfg.Feeds {
		log.Printf("Processing: %s", key)

		opts.Feed = feed
		entries, err := commafeed_util.GetEntries(ctx, c, opts)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}

		// Merge the new list of entries into the existing feed
		if len(mergedEntries) == 0 {
			mergedEntries = entries.Entries
		} else {
			mergedEntries = commafeed_util.MergeEntries(mergedEntries, entries.Entries)
		}

		// If we already have MaxEntries then we want to start tracking a `NewerThan` value
		// which will allow commafeed to return only the entries that will make the cut for the feed
		if len(mergedEntries) > config.Cfg.MaxEntries {
			mergedEntries = mergedEntries[:config.Cfg.MaxEntries]
			newerThan, err := time.Parse(time.RFC3339, mergedEntries[len(mergedEntries)-1].Date)
			if err != nil {
				log.Printf("[MaxEntries] Failed to parse Date: %s: %v", mergedEntries[len(mergedEntries)-1].Date, err)
			} else {
				opts.NewerThan = newerThan
			}
		}
	}

	return mergedEntries
}

func writeOutputFiles(feed *feeds.Feed) {
	if feed == nil {
		panic("feed is nil")
	}

	wroteOne := false

	if config.Cfg.Output.RssFile != "" {
		wroteOne = true
		if rss, err := feed.ToRss(); err == nil {
			fn := filepath.Join(config.Cfg.Output.Webroot, config.Cfg.Output.RssFile)
			if err = os.WriteFile(fn, []byte(rss), 0644); err != nil {
				log.Printf("ERROR: Writting RSS: %v", err)
			}
		}
	}

	if config.Cfg.Output.AtomFile != "" {
		wroteOne = true
		if atom, err := feed.ToAtom(); err == nil {
			fn := filepath.Join(config.Cfg.Output.Webroot, config.Cfg.Output.AtomFile)
			if err = os.WriteFile(fn, []byte(atom), 0644); err != nil {
				log.Printf("ERROR: Writting Atom: %v", err)
			}
		}
	}

	if config.Cfg.Output.JsonFile != "" {
		wroteOne = true
		if js, err := feed.ToJSON(); err == nil {
			fn := filepath.Join(config.Cfg.Output.Webroot, config.Cfg.Output.JsonFile)
			if err = os.WriteFile(fn, []byte(js), 0644); err != nil {
				log.Printf("ERROR: Writting JSON: %v", err)
			}
		}
	}

	if !wroteOne {
		log.Println("No output files were configured, nothing was written")
	}
}

func Generate(forceRebuild bool) bool {
	entries := getFeedEntries()
	if len(entries) == 0 {
		log.Println("No entries, likely an error so won't be generating anything")
		return false
	}

	oldTopper := enrichment.GetFeedTopper()
	newTopper, _ := enrichment.Enrich(entries[0])

	if !forceRebuild && oldTopper != nil {
		if newTopper.Page.Url == oldTopper.Page.Url {
			log.Println("No updates to feed")
			return false
		}
	}
	if err := enrichment.SetFeedTopper(newTopper); err != nil {
		log.Printf("ERROR: Failed to set feed topper: %v", err)
		return false
	}

	var err error
	var rawFeed []*enrichment.EnrichedData
	feed := createEmptyFeed()

	feed.Updated, err = time.Parse(time.RFC3339, entries[0].Date)
	if err != nil {
		log.Printf("ERROR: [generate] Failed to parse updated date: %s: %v", entries[0].Date, err)
		return false
	}
	feed.Created = feed.Updated

	for _, e := range entries {
		enriched, err := enrichment.Enrich(e)
		if err != nil {
			log.Printf("ERROR: Failed to enrich: %v", err)
			continue
		}
		feed.Items = append(feed.Items, generateFeedItem(enriched))
		rawFeed = append(rawFeed, enriched)
	}

	if err := writeRawFeed(rawFeed); err != nil {
		log.Printf("ERROR: Writing raw feed: %v", err)
		return false
	}

	writeOutputFiles(feed)
	return true
}

func writeRawFeed(rawFeed []*enrichment.EnrichedData) error {
	if config.Cfg.Output.RawFile == "" {
		return errors.New("raw_file not set")
	}

	fp, err := os.OpenFile(config.Cfg.Output.RawFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = fp.Close() }()

	if err = json.NewEncoder(fp).Encode(rawFeed); err != nil {
		return err
	}

	return nil
}
