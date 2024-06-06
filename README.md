# zerodayfans
Source for the 0dayfans.com which is a Feed aggregator based on a [CommaFeed](https://github.com/athou/commafeed) backend.

Multiple Commafeed Feeds (and Categories) can be configured with simple match based filters applied to include or exclude feed entries from the final feed.

Once the final feed has been calculated, the ./templates folder's *tmpl and *.html template files will be parsed. The *.html files will be executed to write their respective .html file to the webroot acting as a simple site generator.

In theory you could use this for yourself, but honestly, its pretty opinionated and hacked together specifically for my needs to work, not necessarily to be maintained.