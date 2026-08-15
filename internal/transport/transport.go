package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func DefaultSocket() string {
	uid := os.Getuid()
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return filepath.Join(dir, "vanta.sock")
	}
	return fmt.Sprintf("/tmp/vanta-%d/vanta.sock", uid)
}

type Client struct {
	httpc *http.Client
	token string
	base  string
}

func New(socket, token string) *Client {
	return &Client{
		token: token,
		base:  "http://localhost",
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

func (c *Client) resolveURL(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return c.base + url
}

func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, c.resolveURL(url), nil)
	if err != nil {
		return nil, err
	}
	return c.Do(r)
}

func (c *Client) Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.resolveURL(url), body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", contentType)
	return c.Do(r)
}

type SSEEvent struct {
	ID    string
	Event string
	Data  string
}

func ReadSSE(ctx context.Context, r io.Reader) <-chan SSEEvent {
	ch := make(chan SSEEvent)
	go func() {
		defer close(ch)
		buf := make([]byte, 0, 4096)
		tmp := make([]byte, 1024)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			n, err := r.Read(tmp)
			if n > 0 {
				buf = append(buf, tmp[:n]...)
				for {
					events, rest := parseSSE(buf)
					buf = rest
					if len(events) == 0 {
						break
					}
					for _, ev := range events {
						select {
						case ch <- ev:
						case <-ctx.Done():
							return
						}
					}
				}
			}
			if err != nil {
				return
			}
		}
	}()
	return ch
}

func parseSSE(data []byte) ([]SSEEvent, []byte) {
	var events []SSEEvent

	for {
		idx := bytes.Index(data, []byte("\n\n"))
		if idx < 0 {
			break
		}
		block := data[:idx]
		data = data[idx+2:]

		ev := parseSSEBlock(block)
		if ev.Data != "" || ev.Event != "" {
			events = append(events, ev)
		}
	}

	return events, data
}

func parseSSEBlock(block []byte) SSEEvent {
	var ev SSEEvent
	var dataLines [][]byte
	lines := bytes.Split(block, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimRight(line, "\r")
		if bytes.HasPrefix(line, []byte("id:")) {
			ev.ID = strings.TrimSpace(string(bytes.TrimPrefix(line, []byte("id:"))))
		} else if bytes.HasPrefix(line, []byte("event:")) {
			ev.Event = strings.TrimSpace(string(bytes.TrimPrefix(line, []byte("event:"))))
		} else if bytes.HasPrefix(line, []byte("data:")) {
			dataLines = append(dataLines, bytes.TrimPrefix(line, []byte("data:")))
		}
	}
	if len(dataLines) > 0 {
		ev.Data = strings.TrimSpace(string(bytes.Join(dataLines, []byte("\n"))))
	}
	return ev
}

func WriteSSE(w http.ResponseWriter, id int64, event string, data any) {
	fmt.Fprintf(w, "id: %d\n", id)
	if event != "" {
		fmt.Fprintf(w, "event: %s\n", event)
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", jsonBytes)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
