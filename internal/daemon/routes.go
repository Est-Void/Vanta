package daemon

import (
	"net/http"

	"github.com/Est-Void/Vanta/api"
)

func newRouter(clients map[string][]api.AuthScope) http.Handler {
	mux := http.NewServeMux()
	authed := requireAuth(clients, mux)

	mux.HandleFunc("GET /v1/status", handleStatus)

	mux.Handle("POST /v1/session",
		requireScope(api.ScopeAgent, http.HandlerFunc(handleCreateSession)))
	mux.Handle("POST /v1/session/{id}/message",
		requireScope(api.ScopeAgent, http.HandlerFunc(handleSendMessage)))

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

func handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req api.CreateSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, api.ErrBadRequest, "invalid json")
		return
	}
	writeError(w, api.ErrInternal, "not implemented")
}

func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, api.ErrBadRequest, "missing session id")
		return
	}
	var req api.SendMessage
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, api.ErrBadRequest, "invalid json")
		return
	}
	writeError(w, api.ErrInternal, "not implemented")
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
	writeError(w, api.ErrInternal, "not implemented")
}
