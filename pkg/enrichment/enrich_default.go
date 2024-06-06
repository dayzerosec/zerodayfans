package enrichment

import (
	"bytes"
	"github.com/JohannesKaufmann/html-to-markdown"
	"github.com/adampresley/gofavigrab/parser"
	"github.com/astappiev/microdata"
	"github.com/markusmobius/go-trafilatura"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

func DefaultEnricher(targetUrl string) (*EnrichedData, error) {
	enriched := &EnrichedData{}
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	req, _ := http.NewRequest(http.MethodGet, targetUrl, nil)
	req.Header.Set("User-Agent", "Googlebot/2.1 (+http://www.google.com/bot.html)")

	// Fetch OpenGraph data
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	microData, err := microdata.ParseHTML(bytes.NewReader(data), resp.Header.Get("Content-Type"), targetUrl)
	if err == nil {
		enriched.WithMicrodata(microData)
	}

	ogData, err := ParseOG(data)
	enriched.WithOpenGraph(ogData)

	if enriched.Publisher.Image == "" {
		htmlParser := parser.NewHTMLParser(string(data))
		enriched.Publisher.Image, _ = htmlParser.GetFaviconURL()
		log.Println("Favicon URL:", enriched.Publisher.Image)
	}

	enriched.Page.Content, _ = parseContentAsMd(targetUrl, data)
	enriched.Page.Content = html.UnescapeString(enriched.Page.Content)

	return enriched, nil
}

func parseContentAsMd(targetUrl string, data []byte) (string, error) {
	u, _ := url.Parse(targetUrl)
	extract, err := trafilaturaExtract(data)
	if err != nil {
		return "", err
	}
	converter := md.NewConverter(u.Host, true, nil)

	var contentHtml bytes.Buffer
	err = html.Render(&contentHtml, extract.ContentNode)
	if err != nil {
		return "", err
	}

	buf, err := converter.ConvertReader(&contentHtml)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func trafilaturaExtract(bs []byte) (*trafilatura.ExtractResult, error) {
	var opts trafilatura.Options
	opts.FallbackCandidates = &trafilatura.FallbackConfig{}
	opts.TargetLanguage = ""
	opts.ExcludeComments = false
	opts.ExcludeTables = false
	opts.IncludeImages = false
	opts.IncludeLinks = false
	opts.Deduplicate = true
	opts.HasEssentialMetadata = false
	opts.EnableLog = true

	return trafilatura.Extract(bytes.NewReader(bs), opts)
}
