package enrichment

import (
	"github.com/astappiev/microdata"
	"github.com/julianshen/og"
	"strings"
)

func (d *EnrichedData) WithMicrodata(data *microdata.Microdata) {
	for _, item := range data.Items {
		types := strings.ToLower(strings.Join(item.Types, ", "))
		if strings.Contains(types, "article") || strings.Contains(types, "blogpost") {
			for k, v := range item.Properties {
				if len(v) == 0 {
					continue
				}

				switch k {
				case "url":
					if d.Page.Url == "" {
						d.Page.Url = v[0].(string)
					}
				case "headline":
					if d.Page.Title == "" {
						d.Page.Title = v[0].(string)
					}
				case "abstract":
					fallthrough
				case "description":
					if d.Page.Description == "" {
						d.Page.Description = v[0].(string)
					}
				case "author":
					if pubItem, ok := v[0].(*microdata.Item); ok {
						pub := microdataParsePublisher(pubItem)
						if d.Author.Name == "" {
							d.Author.Name = pub.Name
						}
						if d.Author.Url == "" {
							d.Author.Url = pub.Url
						}
					}
					if str, ok := v[0].(string); ok {
						if d.Author.Name == "" {
							d.Author.Name = str
						}
					}
				case "publisher":
					if pubItem, ok := v[0].(*microdata.Item); ok {
						pub := microdataParsePublisher(pubItem)
						if d.Publisher.Name == "" {
							d.Publisher.Name = pub.Name
						}
						if d.Publisher.Url == "" {
							d.Publisher.Url = pub.Url
						}
						if d.Publisher.Image == "" {
							d.Publisher.Image = pub.Image
						}
					}
					if str, ok := v[0].(string); ok {
						if d.Publisher.Name == "" {
							d.Publisher.Name = str
						}
					}
				case "image":
					d.Page.Image = microdataGetImageUrl(v)
				}
			}
		}

		if strings.Contains(types, "website") {
			for k, v := range item.Properties {
				switch k {
				case "url":
					if d.Publisher.Url == "" {
						d.Publisher.Url = v[0].(string)
					}
				case "name":
					if d.Publisher.Name == "" {
						d.Publisher.Name = v[0].(string)
					}
				case "thumbnailUrl":
					if d.Publisher.Image == "" {
						d.Publisher.Image = v[0].(string)
					}
				case "thumbnail":
					if d.Publisher.Image == "" {
						d.Publisher.Image = microdataGetImageUrl(v)
					}
				case "logo":
					if d.Publisher.Image == "" {
						d.Publisher.Image = microdataGetImageUrl(v)
					}
				case "image":
					// Image will always take precedence for the "image" value so don't check if its empty
					d.Publisher.Image = microdataGetImageUrl(v)
				}

			}
		}

	}
}

func (d *EnrichedData) WithTwitterCard(card *og.TwitterCard) {
	if d.Page.Title == "" {
		d.Page.Title = card.Title
	}
	if d.Page.Description == "" {
		d.Page.Description = card.Description
	}
	if d.Page.Image == "" {
		d.Page.Image = card.Image
	}
	if d.Page.Url == "" {
		d.Page.Url = card.Url
	}
}

func (d *EnrichedData) WithOpenGraph(data *og.PageInfo) {
	d.WithTwitterCard(data.Twitter)

	if d.Page.Url == "" {
		d.Page.Url = data.Url
	}

	if d.Page.Title == "" {
		d.Page.Title = data.Title
	}

	if d.Page.Description == "" {
		d.Page.Description = data.Description
	}

	if d.Page.Image == "" {
		if len(data.Images) > 0 {
			d.Page.Image = data.Images[0].Url
		}
	}

	if d.Publisher.Name == "" {
		d.Publisher.Name = data.SiteName
	}

	// Publisher Url/Image are not part of opengraph
	// Authorship information is not part of opengraph

}
