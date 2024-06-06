package main

import (
	"context"
	"flag"
	"github.com/dayzerosec/zerodayfans/pkg/commafeed_util"
	"github.com/dayzerosec/zerodayfans/pkg/config"
	"github.com/dayzerosec/zerodayfans/pkg/feedgen"
	"github.com/dayzerosec/zerodayfans/pkg/sitegen"
	"log"
	"os"
)

func main() {
	configFn := flag.String("config", "", "Location of the configuration file to use")
	flag.Parse()

	if *configFn == "" {
		flag.Usage()
		return
	}

	if err := config.Load(*configFn); err != nil {
		panic(err)
	}

	forceRebuild := os.Getenv("ZF_REBUILD") != ""

	if !feedgen.Generate(forceRebuild) {
		// Nothing new was generated
		return
	}

	if val, ok := config.Cfg.Feeds["__test"]; ok {
		doTestFeed(val)
		log.Printf("DON'T FORGET TO REMOVE FEEDS.__test FROM CONFIG")
		return
	}

	if err := sitegen.Generate(); err != nil {
		log.Fatalf("Error generating site: %v", err)
	}

}

func doTestFeed(val config.FeedConfig) {
	log.Printf("Found a __test feed: %v", val)
	ctx := context.Background()
	c, _ := config.Cfg.Commafeed.Client()
	opts := commafeed_util.GetEntriesOptions{
		Feed:       val,
		MaxEntries: config.Cfg.MaxEntries,
	}
	entries, err := commafeed_util.GetEntries(ctx, c, opts)
	if err != nil {
		log.Fatalf("Error getting entries: %v", err)
	}

	entries.Entries, err = commafeed_util.FilterEntries(val, entries)
	if err != nil {
		log.Fatalf("Error filtering entries: %v", err)
	}

	commafeed_util.PrintEntries(entries)

	log.Printf("DON'T FORGET TO REMOVE FEEDS.__test FROM CONFIG")
	return
}
