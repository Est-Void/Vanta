package transport

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"
)

type Client struct {
	httpc *http.Client
	token string
}

func New(socket, token string) *Client {
	return &Client{
		token: token,
		httpc: &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.DialTimeout("unix", socket, 2*time.Second)
				},
			},
		},
	}
}

func (c *Client) Do(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", "Bearer "+c.token)
	return c.httpc.Do(r)
}

func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

func (c *Client) Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", contentType)
	return c.Do(r)
}
