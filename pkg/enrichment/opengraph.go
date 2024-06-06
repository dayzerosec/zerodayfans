package enrichment

import (
	"github.com/julianshen/og"
)

func ParseOG(data []byte) (*og.PageInfo, error) {
	out := &og.PageInfo{}
	err := og.GetPageDataFromHtml(data, out)
	return out, err
}
