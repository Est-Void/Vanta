package transport

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultSocket_XDG(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", dir)

	got := DefaultSocket()
	want := filepath.Join(dir, "vanta.sock")
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
	}
}

func TestDefaultSocket_Fallback(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "")

	got := DefaultSocket()
	if filepath.Base(got) != "vanta.sock" {
		t.Errorf("expected path ending with vanta.sock, got %s", got)
	}
}

func TestClient_Do_AddsAuthHeader(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{
		token: "test-token",
		httpc: srv.Client(),
	}

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	_, err := c.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "Bearer test-token"
	if gotHeader != want {
		t.Errorf("expected Authorization=%q, got %q", want, gotHeader)
	}
}

func TestClient_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/test" {
			t.Errorf("expected /test, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := &Client{
		token: "tok",
		httpc: srv.Client(),
	}

	resp, err := c.Get(t.Context(), srv.URL+"/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClient_Post(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		gotBody = string(buf[:n])
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{
		token: "tok",
		httpc: srv.Client(),
	}

	body, _ := json.Marshal(map[string]string{"key": "value"})
	resp, err := c.Post(t.Context(), srv.URL+"/test", "application/json", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var got map[string]string
	if err := json.Unmarshal([]byte(gotBody), &got); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("expected key=value, got %v", got)
	}
}

func TestClient_UnixSocket(t *testing.T) {
	dir := t.TempDir()
	sock := filepath.Join(dir, "test.sock")

	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatalf("failed to create socket: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 4096)
				c.Read(buf)
				resp := "HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok"
				c.Write([]byte(resp))
			}(conn)
		}
	}()

	c := New(sock, "tok")
	resp, err := c.Get(t.Context(), "http://localhost/test")
	if err != nil {
		t.Fatalf("failed to connect to unix socket: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestNew(t *testing.T) {
	c := New("/tmp/test.sock", "my-token")
	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.token != "my-token" {
		t.Errorf("expected token 'my-token', got %q", c.token)
	}
}

func TestParseSSE_SingleEvent(t *testing.T) {
	data := []byte("event: token\ndata: {\"text\":\"hello\"}\n\n")
	events, rest := parseSSE(data)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Event != "token" {
		t.Errorf("expected event 'token', got %q", events[0].Event)
	}
	if events[0].Data != `{"text":"hello"}` {
		t.Errorf("unexpected data: %q", events[0].Data)
	}
	if len(rest) != 0 {
		t.Errorf("expected empty rest, got %q", rest)
	}
}

func TestParseSSE_MultipleEvents(t *testing.T) {
	data := []byte("event: token\ndata: hello\n\nevent: done\ndata: {}\n\n")
	events, rest := parseSSE(data)

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Event != "token" {
		t.Errorf("expected first event 'token', got %q", events[0].Event)
	}
	if events[1].Event != "done" {
		t.Errorf("expected second event 'done', got %q", events[1].Event)
	}
	if len(rest) != 0 {
		t.Errorf("expected empty rest, got %q", rest)
	}
}

func TestParseSSE_IncompleteEvent(t *testing.T) {
	data := []byte("event: token\ndata: hello\n")
	events, rest := parseSSE(data)

	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
	if len(rest) != len(data) {
		t.Errorf("expected rest to be original data, got %d bytes", len(rest))
	}
}

func TestParseSSE_DataOnly(t *testing.T) {
	data := []byte("data: hello\n\n")
	events, _ := parseSSE(data)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Event != "" {
		t.Errorf("expected empty event type, got %q", events[0].Event)
	}
	if events[0].Data != "hello" {
		t.Errorf("expected data 'hello', got %q", events[0].Data)
	}
}

func TestReadSSE(t *testing.T) {
	sseData := "event: token\ndata: hello\n\nevent: done\ndata: {}\n\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte(sseData))
	}))
	defer srv.Close()

	c := &Client{
		token: "tok",
		httpc: srv.Client(),
	}

	resp, err := c.Get(t.Context(), srv.URL+"/events")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	events := ReadSSE(t.Context(), resp.Body)

	var got []SSEEvent
	for ev := range events {
		got = append(got, ev)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
	if got[0].Event != "token" {
		t.Errorf("expected 'token', got %q", got[0].Event)
	}
	if got[1].Event != "done" {
		t.Errorf("expected 'done', got %q", got[1].Event)
	}
}

func TestWriteSSE(t *testing.T) {
	rr := httptest.NewRecorder()

	WriteSSE(rr, 42, "token", map[string]string{"text": "hello"})

	body := rr.Body.String()
	if !strings.Contains(body, "event: token") {
		t.Errorf("missing event type in output: %s", body)
	}
	if !strings.Contains(body, "data: ") {
		t.Errorf("missing data in output: %s", body)
	}
	if !strings.Contains(body, "\n\n") {
		t.Errorf("missing double newline in output: %s", body)
	}
	if !strings.Contains(body, "id: 42") {
		t.Errorf("missing id in output: %s", body)
	}
}

func TestParseSSE_WithID(t *testing.T) {
	data := []byte("id: 123\nevent: token\ndata: hello\n\n")
	events, _ := parseSSE(data)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ID != "123" {
		t.Errorf("expected id '123', got %q", events[0].ID)
	}
	if events[0].Event != "token" {
		t.Errorf("expected event 'token', got %q", events[0].Event)
	}
}
