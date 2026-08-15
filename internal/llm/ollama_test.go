package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOllama_Chat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("expected /api/chat, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req ollamaChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "llama3" {
			t.Errorf("expected model llama3, got %s", req.Model)
		}
		if len(req.Messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(req.Messages))
		}
		if req.Messages[0].Content != "hello" {
			t.Errorf("expected content 'hello', got %s", req.Messages[0].Content)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ollamaChatResponse{
			Message: struct {
				Content string `json:"content"`
			}{Content: "hi there"},
			Done:            true,
			PromptEvalCount: 10,
			EvalCount:       5,
		})
	}))
	defer srv.Close()

	p := NewOllama(srv.URL)
	resp, err := p.Chat(context.Background(), Request{
		Model:    "llama3",
		Messages: []Message{{Role: RoleUser, Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Content != "hi there" {
		t.Errorf("expected 'hi there', got %q", resp.Content)
	}
	if resp.Usage.PromptTokens != 10 {
		t.Errorf("expected 10 prompt tokens, got %d", resp.Usage.PromptTokens)
	}
	if resp.Usage.CompletionTokens != 5 {
		t.Errorf("expected 5 completion tokens, got %d", resp.Usage.CompletionTokens)
	}
	if resp.FinishReason != FinishStop {
		t.Errorf("expected finish stop, got %s", resp.FinishReason)
	}
}

func TestOllama_ChatStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")

		chunks := []ollamaStreamChunk{
			{Message: struct {
				Content string `json:"content"`
			}{Content: "hel"}, Done: false},
			{Message: struct {
				Content string `json:"content"`
			}{Content: "lo"}, Done: false},
			{Message: struct {
				Content string `json:"content"`
			}{Content: ""}, Done: true, PromptEvalCount: 5, EvalCount: 3},
		}

		enc := json.NewEncoder(w)
		for _, c := range chunks {
			enc.Encode(c)
		}
	}))
	defer srv.Close()

	p := NewOllama(srv.URL)
	ch, err := p.ChatStream(context.Background(), Request{
		Model:    "llama3",
		Messages: []Message{{Role: RoleUser, Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var content strings.Builder
	var gotUsage *Usage
	for delta := range ch {
		content.WriteString(delta.Content)
		if delta.Usage != nil {
			gotUsage = delta.Usage
		}
	}

	if content.String() != "hello" {
		t.Errorf("expected 'hello', got %q", content.String())
	}
	if gotUsage == nil {
		t.Error("expected usage in final chunk")
	}
}

func TestOllama_Models(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("expected /api/tags, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ollamaModelsResponse{
			Models: []struct {
				Name string `json:"name"`
			}{
				{Name: "llama3"},
				{Name: "mistral"},
			},
		})
	}))
	defer srv.Close()

	p := NewOllama(srv.URL)
	models, err := p.Models(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
	if models[0] != "llama3" {
		t.Errorf("expected llama3, got %s", models[0])
	}
	if models[1] != "mistral" {
		t.Errorf("expected mistral, got %s", models[1])
	}
}

func TestOllama_Chat_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	p := NewOllama(srv.URL)
	_, err := p.Chat(context.Background(), Request{
		Model:    "llama3",
		Messages: []Message{{Role: RoleUser, Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 in error, got: %v", err)
	}
}

func TestOllama_DefaultBaseURL(t *testing.T) {
	p := NewOllama("")
	if p.baseURL != defaultBaseURL {
		t.Errorf("expected %s, got %s", defaultBaseURL, p.baseURL)
	}
}

func TestNewOllama(t *testing.T) {
	p := NewOllama("http://custom:8080")
	if p.baseURL != "http://custom:8080" {
		t.Errorf("expected http://custom:8080, got %s", p.baseURL)
	}
}
