package commafeed_util

import (
	"github.com/dayzerosec/zerodayfans/pkg/commafeed"
	"log"
	"time"
)

func PrintEntries(entries *commafeed.Entries) {
	log.Printf("==============================")
	log.Printf("Name: %s (%s)", entries.Name, *entries.FeedLink)
	log.Printf("==============================")

	for _, e := range entries.Entries {
		log.Printf("%s", e.Title)
		log.Printf("\tCats:\t%v", *e.Categories)
		log.Printf("\tLink:\t%s", e.Url)
		log.Printf("\tAuth:\t%s", *e.Author)
		log.Printf("\tDate:\t%s", time.UnixMilli(int64(e.Date)))
		log.Printf("------------------------------")
	}
}
