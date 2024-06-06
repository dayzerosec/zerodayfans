package config

import (
	"context"
	"github.com/dayzerosec/zerodayfans/pkg/commafeed"
	"github.com/labstack/gommon/log"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type CommafeedConfig struct {
	BaseUrl string `toml:"base_url"`
	ApiKey  string `toml:"api_key"`
	// Username is optional and only required for tracking categories as that requires a user session
	Username string `toml:"username"`
	// Password is optional and only required for tracking categories as that requires a user session
	Password string `toml:"password"`

	requestEditor commafeed.RequestEditorFn
	client        *commafeed.ClientWithResponses
}

func (c *CommafeedConfig) SetRequestEditorFn(fn commafeed.RequestEditorFn) {
	c.requestEditor = fn
	if _, err := c.generateClient(); err != nil {
		log.Error("Error creating Commafeed client: ", err)
	}
}

func (c *CommafeedConfig) generateClient() (*commafeed.ClientWithResponses, error) {
	var opts []commafeed.ClientOption

	if c.ApiKey != "" {
		opts = append(opts, commafeed.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			if c.ApiKey != "" {
				q := req.URL.Query()
				q.Add("apiKey", c.ApiKey)
				req.URL.RawQuery = q.Encode()
			}
			return nil
		}))
	}

	if c.requestEditor != nil {
		opts = append(opts, commafeed.WithRequestEditorFn(c.requestEditor))
	}

	jar, _ := cookiejar.New(&cookiejar.Options{})
	baseHttpClient := http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}
	opts = append(opts, commafeed.WithHTTPClient(&baseHttpClient))

	client, err := commafeed.NewClientWithResponses(
		c.BaseUrl,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	c.client = client
	return client, nil
}

func (c *CommafeedConfig) Client() (*commafeed.ClientWithResponses, error) {
	if c.client == nil {
		client, err := c.generateClient()
		if err != nil {
			return nil, err
		}
		c.client = client
	}
	return c.client, nil
}
