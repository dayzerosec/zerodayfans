package enrichment

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func init() {
	RegisterEnricher("hackerone.com", EnrichHackerOneReport)
}

type H1Report struct {
	Id                       int        `json:"id"`
	Url                      string     `json:"url"`
	Title                    string     `json:"title"`
	SeverityRating           string     `json:"severity_rating"`
	Reporter                 H1Reporter `json:"reporter"`
	Team                     H1Team     `json:"team"`
	VulnerabilityInformation string     `json:"vulnerability_information"`
}

type H1Reporter struct {
	Username        string             `json:"username"`
	Url             string             `json:"url"`
	ProfilePictures ProfilePictureUrls `json:"profile_picture_urls"`
}

type H1Team struct {
	Id              int                `json:"id"`
	Url             string             `json:"url"`
	Handle          string             `json:"handle"`
	ProfilePictures ProfilePictureUrls `json:"profile_picture_urls"`
}

type ProfilePictureUrls struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

func (p ProfilePictureUrls) Best() string {
	if p.Large != "" {
		return p.Large
	}
	if p.Medium != "" {
		return p.Medium
	}
	if p.Small != "" {
		return p.Small
	}
	return ""
}

func EnrichHackerOneReport(targetUrl string) (*EnrichedData, error) {
	reportRegex := regexp.MustCompile(`^https?://hackerone\.com/reports/(\d+)$`)
	// Get report id from regex
	reportIdMatches := reportRegex.FindStringSubmatch(targetUrl)
	if len(reportIdMatches) <= 1 {
		return nil, ErrCantEnrich
	}
	reportId := reportIdMatches[1]
	reportJsonUrl := fmt.Sprintf("https://hackerone.com/reports/%s.json", reportId)

	res, err := http.Get(reportJsonUrl)
	if err != nil {
		log.Printf("[ERROR] Failed to get HackerOne report: %v", err)
		return nil, ErrCantEnrich
	}
	defer func() { _ = res.Body.Close() }()

	report := &H1Report{}
	if err := json.NewDecoder(res.Body).Decode(report); err != nil {
		log.Printf("[ERROR] Failed to decode HackerOne report JSON: %v", err)
		return nil, ErrCantEnrich
	}

	enriched := &EnrichedData{}
	enriched.Publisher.Name = fmt.Sprintf("%s - HackerOne", report.Team.Handle)
	enriched.Publisher.Url = report.Team.Url

	// I want to use this but its a timed S3 link, so I'd have to come up with a cache strategy for images
	//enriched.Publisher.Image = report.Team.ProfilePictures.Best()
	enriched.Publisher.Image = "/static/h1_mark_black.png"

	enriched.Author.Name = report.Reporter.Username
	if strings.HasPrefix(report.Reporter.Url, "/") {
		enriched.Author.Url = fmt.Sprintf("https://hackerone.com%s", report.Reporter.Url)
	} else {
		enriched.Author.Url = report.Reporter.Url
	}

	enriched.Page.Title = report.Title
	enriched.Page.Url = report.Url
	enriched.Page.Content = report.VulnerabilityInformation

	return enriched, nil
}
