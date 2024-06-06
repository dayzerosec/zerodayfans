package enrichment

import (
	"github.com/astappiev/microdata"
)

type PublisherMicrodata struct {
	Name  string
	Url   string
	Image string
}

func microdataGetImageUrl(data microdata.ValueList) string {
	// images can be an ImageObject or a Url so have to handle both cases
	for _, v := range data {
		if item, ok := v.(*microdata.Item); ok {
			for k, v := range item.Properties {
				if len(v) == 0 {
					continue
				}
				switch k {
				case "contentUrl":
					fallthrough
				case "url":
					return v[0].(string)
				}
			}
		}
		if url, ok := v.(string); ok {
			return url
		}
	}
	return ""
}

func microdataParsePublisher(data *microdata.Item) PublisherMicrodata {
	publisher := PublisherMicrodata{}

	for k, v := range data.Properties {
		if len(v) == 0 {
			continue
		}

		switch k {
		case "url":
			publisher.Url = v[0].(string)
		case "name":
			publisher.Name = v[0].(string)
		case "logo":
			publisher.Image = microdataGetImageUrl(v)
		}
	}

	return publisher
}
