package config

type FeedType string

const (
	FeedTypeCategory FeedType = "category"
	FeedTypeFeed     FeedType = "feed"
)

type FeedConfig struct {
	// Type is used to determine which API endpoint is used to retrieve posts from commafeed
	Type FeedType `toml:"type"`
	// Id is the commafeed id of the feed or category to be retrieved
	ID string `toml:"id"`
	// UrlRewrite should be an array of TWO strings, first is a regex match, second is a regex replacement. Useful
	// for fixing commafeed_util that have malformed or broken URLs.
	UrlRewrite []string              `toml:"url_rewrite"`
	Filters    map[string]FeedFilter `toml:"filters"`
	DailyLimit int                   `toml:"daily_limit"`
}

type FilterType string

const (
	FilterTypeMatchOne FilterType = "match_one"
	FilterTypeMatchAll FilterType = "match_all"
)

type FeedFilter struct {
	Field         string     `toml:"field"`
	Type          FilterType `toml:"type"`
	CaseSensitive bool       `toml:"case_sensitive"`
	Values        []string   `toml:"values"`
	Negate        bool       `toml:"negate"`
}
