package main

import (
	"context"
	"github.com/dayzerosec/zerodayfans/pkg/commafeed_util"
	"github.com/dayzerosec/zerodayfans/pkg/config"
	"log"
	"time"
)

func main() {
	if err := config.Load(".local/config.toml"); err != nil {
		panic(err)
	}

	ctx := context.Background()
	c, _ := config.Cfg.Commafeed.Client()
	opts := commafeed_util.GetEntriesOptions{
		Feed:       config.FeedConfig{},
		NewerThan:  time.Time{},
		MaxEntries: config.Cfg.MaxEntries,
	}

	for key, feed := range config.Cfg.Feeds {
		opts.Feed = feed
		entries, err := commafeed_util.GetEntries(ctx, c, opts)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}

		log.Println(key)
		if len(entries.Entries) == 0 {
			log.Println("\tNo entries found")
			continue
		}
		if date, err := time.Parse(time.RFC3339, entries.Entries[0].Date); err == nil {
			log.Printf("\tDate: %v", date)
		} else {
			log.Printf("\t[!] Error parsing date: %v", err)
		}

		for _, entry := range entries.Entries {
			_, err := time.Parse(time.RFC3339, entry.Date)
			if err != nil {
				log.Printf("\t[!] Error parsing date: %v", err)
			}
		}

	}
}
