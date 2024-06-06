package enrichment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type monorailComment struct {
	Timestamp int    `json:"timestamp"`
	Content   string `json:"content"`
}

type monorailIssue struct {
	ProjectName string `json:"projectName"`
	LocalId     int    `json:"localId"`
	Summary     string `json:"summary"`

	StatusRef struct {
		Status string `json:"status"`
	} `json:"statusRef"`

	OwnerRef    monorailUserRef   `json:"ownerRef"`
	CCRefs      []monorailUserRef `json:"ccRefs"`
	ReporterRef monorailUserRef   `json:"reporterRef"`

	LabelRefs []struct {
		Label string `json:"label"`
	} `json:"labelRefs"`

	OpenedTimestamp   int `json:"openedTimestamp"`
	ClosedTimestamp   int `json:"closedTimestamp"`
	ModifiedTimestamp int `json:"modifiedTimestamp"`
}

type monorailUserRef struct {
	UserId      string `json:"userId"`
	DisplayName string `json:"displayName"`
}

type monorailReqBody struct {
	IssueRef struct {
		LocalId     int    `json:"localId"`
		ProjectName string `json:"projectName"`
	} `json:"issueRef"`
}

func init() {
	RegisterEnricher("bugs.chromium.org", MonorailEnrich)
}

func monorailGetXsrfToken() (token string, err error) {
	// This is usually embedded into page JS, I'm just going to fetch a known page with it and hope it isn't scoped
	u := "https://bugs.chromium.org/p/project-zero/issues/list"

	var res *http.Response
	res, err = http.Get(u)
	if err != nil {
		return
	}
	defer func() { _ = res.Body.Close() }()

	bs, _ := io.ReadAll(res.Body)

	if !bytes.Contains(bs, []byte("'token': '")) {
		err = fmt.Errorf("no token found")
		return
	}

	bs = bs[bytes.Index(bs, []byte("'token': '"))+10:]
	bs = bs[:bytes.Index(bs, []byte("'"))]
	token = string(bs)

	return
}

func monorailListComments(xsrfToken, project string, localId int) ([]monorailComment, error) {
	var reqBody monorailReqBody
	reqBody.IssueRef.LocalId = localId
	reqBody.IssueRef.ProjectName = project

	u := "https://bugs.chromium.org/prpc/monorail.Issues/ListComments"

	bs, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(bs))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Xsrf-Token", xsrfToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = res.Body.Close() }()

	bs, _ = io.ReadAll(res.Body)
	bs = bytes.TrimSpace(bs[bytes.Index(bs, []byte("\n")):])

	respBody := &struct {
		Comments []monorailComment `json:"comments"`
	}{}

	err = json.Unmarshal(bs, respBody)
	if err != nil {
		return nil, err
	}

	return respBody.Comments, nil
}
func monorailGetIssue(xsrfToken, project string, localId int) (*monorailIssue, error) {
	var reqBody monorailReqBody
	reqBody.IssueRef.LocalId = localId
	reqBody.IssueRef.ProjectName = project

	u := "https://bugs.chromium.org/prpc/monorail.Issues/GetIssue"

	bs, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(bs))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Xsrf-Token", xsrfToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = res.Body.Close() }()

	bs, _ = io.ReadAll(res.Body)
	bs = bytes.TrimSpace(bs[bytes.Index(bs, []byte("\n")):])

	respBody := &struct {
		Issue *monorailIssue `json:"issue"`
	}{}

	err = json.Unmarshal(bs, respBody)
	if err != nil {
		return nil, err
	}

	return respBody.Issue, nil
}

func MonorailEnrich(target string) (*EnrichedData, error) {
	u, _ := url.Parse(target)

	path := strings.TrimLeft(u.Path, "/")
	pathParts := strings.Split(path, "/")
	if len(pathParts) < 2 {
		return nil, ErrCantEnrich
	}
	projectId := pathParts[1]

	ticketIdInt, err := strconv.Atoi(u.Query().Get("id"))
	if err != nil {
		return nil, ErrCantEnrich
	}

	xsrfToken, err := monorailGetXsrfToken()
	if err != nil {
		log.Printf("[MonorailEnrich] Error getting xsrf token: %v", err)
		return nil, ErrCantEnrich

	}

	iss, err := monorailGetIssue(xsrfToken, projectId, ticketIdInt)
	if err != nil {
		log.Printf("[MonorailEnrich] Error getting issue: %v", err)
		return nil, ErrCantEnrich
	}

	comm, err := monorailListComments(xsrfToken, projectId, ticketIdInt)
	if err != nil {
		log.Printf("[MonorailEnrich] Error listing comments: %v", err)
		return nil, ErrCantEnrich
	}

	projectName := strings.Replace(iss.ProjectName, "-", " ", -1)
	projectName = cases.Title(language.English).String(projectName)

	ed := &EnrichedData{
		Publisher: SiteInfo{
			Name: fmt.Sprintf("%s Bug Tracker", projectName),
			Url:  fmt.Sprintf("https://bugs.chromium.org/p/%s/issues/list", iss.ProjectName),
		},
		Page: PageInfo{
			Title: iss.Summary,
			Url:   target,
		},
		Author: AuthorInfo{
			Name: iss.ReporterRef.DisplayName,
			Url:  fmt.Sprintf("https://bugs.chromium.org/u/%s/updates", iss.ReporterRef.UserId),
		},
	}

	if len(comm) > 0 {
		ed.Page.Content = comm[0].Content
	}

	return ed, nil
}
