package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultBaseURL = "http://localhost:11434"

type Ollama struct {
	baseURL    string
	httpClient *http.Client
}

func NewOllama(baseURL string) *Ollama {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Ollama{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done               bool  `json:"done"`
	PromptEvalCount    int   `json:"prompt_eval_count"`
	EvalCount          int   `json:"eval_count"`
	PromptEvalDuration int64 `json:"prompt_eval_duration"`
	EvalDuration       int64 `json:"eval_duration"`
}

type ollamaStreamChunk struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done            bool `json:"done"`
	PromptEvalCount int  `json:"prompt_eval_count"`
	EvalCount       int  `json:"eval_count"`
}

type ollamaModelsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func (o *Ollama) Chat(ctx context.Context, req Request) (*Response, error) {
	ollamaReq := o.toOllamaRequest(req, false)

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error %d: %s", resp.StatusCode, errBody)
	}

	var ollamaResp ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &Response{
		Content:      ollamaResp.Message.Content,
		Usage:        &Usage{PromptTokens: ollamaResp.PromptEvalCount, CompletionTokens: ollamaResp.EvalCount},
		FinishReason: FinishStop,
	}, nil
}

func (o *Ollama) ChatStream(ctx context.Context, req Request) (<-chan StreamDelta, error) {
	ollamaReq := o.toOllamaRequest(req, true)

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ollama error %d: %s", resp.StatusCode, errBody)
	}

	ch := make(chan StreamDelta)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			var chunk ollamaStreamChunk
			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				continue
			}

			delta := StreamDelta{
				Content: chunk.Message.Content,
			}
			if chunk.Done {
				// TODO: map Ollama finish reasons to FinishTool/FinishLength when tool calls are implemented
				delta.FinishReason = FinishStop
				delta.Usage = &Usage{
					PromptTokens:     chunk.PromptEvalCount,
					CompletionTokens: chunk.EvalCount,
				}
			}

			select {
			case ch <- delta:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

func (o *Ollama) Models(ctx context.Context) ([]string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, o.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error %d: %s", resp.StatusCode, errBody)
	}

	var modelsResp ollamaModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	names := make([]string, len(modelsResp.Models))
	for i, m := range modelsResp.Models {
		names[i] = m.Name
	}
	return names, nil
}

// TODO: map req.Tools to Ollama tools format (POST /api/chat supports "tools" field)
// TODO: map req.MaxTokens to Ollama options.num_predict
func (o *Ollama) toOllamaRequest(req Request, stream bool) ollamaChatRequest {
	msgs := make([]ollamaMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = ollamaMessage{
			Role:    string(m.Role),
			Content: m.Content,
		}
	}
	return ollamaChatRequest{
		Model:    req.Model,
		Messages: msgs,
		Stream:   stream,
	}
}
