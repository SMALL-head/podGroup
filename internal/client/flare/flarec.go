package flare

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	FLARE_BACKEND_URL = "FLARE_BACKEND_URL"
)

type Client struct {
	base       *url.URL
	httpClient *http.Client
}

// NewFlareClient 创建访问flare后端的http客户端，如果URL为空，则使用环境变量FLARE_BACKEND_URL
func NewFlareClient(backendURL string) (*Client, error) {
	if backendURL == "" {
		env, _ := os.LookupEnv(FLARE_BACKEND_URL)
		if env == "" {
			return nil, errors.New("flare backend url is empty")
		}
		backendURL = env
	}

	res := new(Client)

	if u, err := url.ParseRequestURI(backendURL); err != nil {
		return nil, errors.New("flare backend url is invalid")
	} else {
		res.base = u
	}

	c := &http.Client{
		Transport: &http.Transport{},
		Timeout:   3 * time.Second,
	}

	res.httpClient = c

	return res, nil

}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c *Client) NewRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	if path == "" {
		return nil, errors.New("path is empty")
	}

	absPath := c.base.ResolveReference(&url.URL{Path: path})
	req, err := http.NewRequestWithContext(ctx, method, absPath.String(), body)
	if err != nil {
		return nil, err
	}
	return req, nil
}
