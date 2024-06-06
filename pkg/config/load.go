package config

import "github.com/BurntSushi/toml"

var Cfg *Config

type Config struct {
	TemplatesDir    string                `toml:"templates_dir"`
	MaxEntries      int                   `toml:"max_entries"`
	EnrichmentCache CacheConfig           `toml:"enrichment_cache"`
	Commafeed       *CommafeedConfig      `toml:"commafeed"`
	Output          OutputConfig          `toml:"output"`
	Sidebar         SidebarConfig         `toml:"sidebar"`
	Feeds           map[string]FeedConfig `toml:"feeds"`
	metadata        toml.MetaData
}

type CacheConfig struct {
	File  string `toml:"file"`
	Table string `toml:"table"`
}

type SidebarConfig struct {
	Order []string                        `toml:"order"`
	Items map[string]SidebarContentConfig `toml:"items"`
}
type SidebarContentConfig struct {
	Title string     `toml:"title"`
	Links [][]string `toml:"links"`
}

type OutputConfig struct {
	Title       string `toml:"title"`
	Subtitle    string `toml:"subtitle"`
	Url         string `toml:"link"`
	Description string `toml:"description"`
	Author      string `toml:"author"`

	// Webroot will contain the index.html template output along with a copy of ./templates/static
	// and all the output files files will be written under it
	Webroot      string `toml:"webroot"`
	RssFile      string `toml:"rss_file"`
	AtomFile     string `toml:"atom_file"`
	JsonFile     string `toml:"json_file"`
	FeedlistFile string `toml:"feedlist_file"`
	// DefaultFavicon should be relative to the webroot or an absolute URL
	DefaultFavicon string `toml:"default_favicon"`

	// ImageRoot is the directory under webroot to store images
	ImageRoot string `toml:"image_root"`

	// RawFile is the only non-optional output file as it is the cache used when generating the final site
	RawFile string `toml:"raw_file"`
}

func Load(filename string) error {
	Cfg = &Config{}
	md, err := toml.DecodeFile(filename, Cfg)
	if err != nil {
		return err
	}
	Cfg.metadata = md
	return nil
}
