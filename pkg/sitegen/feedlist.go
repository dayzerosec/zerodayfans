package sitegen

import (
	"errors"
	"github.com/dayzerosec/zerodayfans/pkg/commafeed_util"
	"github.com/dayzerosec/zerodayfans/pkg/config"
	"os"
	"path/filepath"
	"sort"
)

func createFeedList() error {
	if config.Cfg.Output.FeedlistFile == "" {
		return nil
	}
	if len(commafeed_util.TrackedFeeds) == 0 {
		return errors.New("empty feedlist")
	}

	sort.SliceStable(commafeed_util.TrackedFeeds, func(i, j int) bool {
		return commafeed_util.TrackedFeeds[i].Name < commafeed_util.TrackedFeeds[j].Name
	})

	fp, err := os.Create(filepath.Join(config.Cfg.Output.Webroot, config.Cfg.Output.FeedlistFile))
	if err != nil {
		return err
	}
	defer func() { _ = fp.Close() }()

	_, _ = fp.WriteString("feed name,link,rss_url")
	for _, feed := range commafeed_util.TrackedFeeds {
		_, _ = fp.WriteString("\n" + feed.Name + "," + feed.Link + "," + feed.RssUrl)
	}
	return nil
}
