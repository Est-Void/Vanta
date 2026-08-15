package daemon

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Est-Void/Vanta/api"
	"github.com/Est-Void/Vanta/internal/llm"
	"github.com/Est-Void/Vanta/internal/transport"
)

func newRouter(d *deps) http.Handler {
	mux := http.NewServeMux()
	authed := requireAuth(d.clients, mux)

	mux.HandleFunc("GET /v1/status", handleStatus)

	mux.Handle("POST /v1/session",
		requireScope(api.ScopeAgent, http.HandlerFunc(d.handleCreateSession)))
	mux.Handle("POST /v1/session/{id}/message",
		requireScope(api.ScopeAgent, http.HandlerFunc(d.handleSendMessage)))

	mux.Handle("POST /v1/attachment",
		requireScope(api.ScopeAgent, http.HandlerFunc(handleUploadAttachment)))

	mux.Handle("POST /v1/terminal",
		requireScope(api.ScopeTerminal, http.HandlerFunc(handleTerminal)))
	mux.Handle("POST /v1/terminal/{id}/input",
		requireScope(api.ScopeTerminal, http.HandlerFunc(handleTerminalInput)))
	mux.Handle("GET /v1/terminal/{id}/stream",
		requireScope(api.ScopeTerminal, http.HandlerFunc(handleTerminalStream)))

	mux.Handle("POST /v1/screen/capture",
		requireScope(api.ScopeScreen, http.HandlerFunc(handleCapture)))

	mux.Handle("POST /v1/input/mouse",
		requireScope(api.ScopeInput, http.HandlerFunc(handleInputMouse)))
	mux.Handle("POST /v1/input/key",
		requireScope(api.ScopeInput, http.HandlerFunc(handleInputKey)))

	mux.Handle("GET /v1/device/info",
		requireScope(api.ScopeDevice, http.HandlerFunc(handleDeviceInfo)))

	mux.Handle("POST /v1/voice/transcribe",
		requireScope(api.ScopeVoice, http.HandlerFunc(handleVoiceTranscribe)))

	mux.HandleFunc("GET /v1/events", handleEvents)

	return authed
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, api.Status{
		Data: map[string]any{"ok": true},
	})
}

func (d *deps) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req api.CreateSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, api.ErrBadRequest, "invalid json")
		return
	}

	id := fmt.Sprintf("sess-%d", time.Now().UnixNano())
	title := req.Title
	if title == "" {
		title = "New session"
	}
	d.sessions.Create(id, title)

	writeJSON(w, http.StatusOK, api.CreateSessionResult{
		Session: api.Session{
			ID:        id,
			Title:     title,
			CreatedAt: time.Now(),
		},
	})
}

func (d *deps) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	if sessionID == "" {
		writeError(w, api.ErrBadRequest, "missing session id")
		return
	}
	var req api.SendMessage
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, api.ErrBadRequest, "invalid json")
		return
	}

	sess := d.sessions.GetOrCreate(sessionID, "Chat")

	if err := d.sessions.AddMessage(sessionID, api.RoleUser, req.Content); err != nil {
		writeError(w, api.ErrNotFound, err.Error())
		return
	}

	history := d.sessions.GetMessages(sessionID)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()
	start := time.Now()

	stream, err := d.llm.ChatStream(ctx, llm.Request{
		Model:    d.model,
		Messages: history,
	})
	if err != nil {
		id := sess.NextEventID()
		transport.WriteSSE(w, id, string(api.EvError), api.ErrorData{
			Error: api.Error{Code: api.ErrInternal, Message: err.Error()},
		})
		transport.WriteSSE(w, sess.NextEventID(), string(api.EvDone), api.DoneData{
			Duration: time.Since(start),
			Err:      &api.Error{Code: api.ErrInternal, Message: err.Error()},
		})
		return
	}

	var fullContent string
	var usage *api.TokenUsage
	defer func() {
		d.sessions.AddMessage(sessionID, api.RoleAgent, fullContent)
		transport.WriteSSE(w, sess.NextEventID(), string(api.EvDone), api.DoneData{
			Duration: time.Since(start),
			Usage:    usage,
		})
	}()

	for delta := range stream {
		if delta.Content != "" {
			fullContent += delta.Content
			transport.WriteSSE(w, sess.NextEventID(), string(api.EvToken), api.TokenData{Text: delta.Content})
		}
		if delta.Usage != nil {
			usage = &api.TokenUsage{
				Prompt:     delta.Usage.PromptTokens,
				Completion: delta.Usage.CompletionTokens,
			}
		}
		if delta.FinishReason != "" {
			break
		}
	}
}

func handleUploadAttachment(w http.ResponseWriter, r *http.Request) {
	writeError(w, api.ErrInternal, "not implemented")
}

func handleTerminal(w http.ResponseWriter, r *http.Request) {
	var req api.CreateTerminalRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, api.ErrBadRequest, "invalid json")
		return
	}
	writeError(w, api.ErrInternal, "not implemented")
}

func handleTerminalInput(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, api.ErrBadRequest, "missing terminal id")
		return
	}
	var req api.WriteInput
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, api.ErrBadRequest, "invalid json")
		return
	}
	writeError(w, api.ErrInternal, "not implemented")
}

func handleTerminalStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, api.ErrBadRequest, "missing terminal id")
		return
	}
	writeError(w, api.ErrInternal, "not implemented")
}

func handleCapture(w http.ResponseWriter, r *http.Request) {
	var req api.CaptureRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, api.ErrBadRequest, "invalid json")
		return
	}
	writeError(w, api.ErrInternal, "not implemented")
}

func handleInputMouse(w http.ResponseWriter, r *http.Request) {
	var req api.MouseEvent
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, api.ErrBadRequest, "invalid json")
		return
	}
	writeError(w, api.ErrInternal, "not implemented")
}

func handleInputKey(w http.ResponseWriter, r *http.Request) {
	var req api.KeyEvent
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, api.ErrBadRequest, "invalid json")
		return
	}
	writeError(w, api.ErrInternal, "not implemented")
}

func handleDeviceInfo(w http.ResponseWriter, r *http.Request) {
	writeError(w, api.ErrInternal, "not implemented")
}

func handleVoiceTranscribe(w http.ResponseWriter, r *http.Request) {
	var req api.TranscribeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, api.ErrBadRequest, "invalid json")
		return
	}
	writeError(w, api.ErrInternal, "not implemented")
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	transport.WriteSSE(w, 1, string(api.EvDone), api.DoneData{})
}
