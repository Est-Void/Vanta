package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Est-Void/Vanta/api"
)

func TestHandleStatus(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/status", nil)
	rr := httptest.NewRecorder()

	handleStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var status api.Status
	if err := json.NewDecoder(rr.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if status.Data["ok"] != true {
		t.Errorf("expected ok=true, got %v", status.Data["ok"])
	}
}

func TestHandleCreateSession_ValidJSON(t *testing.T) {
	d := &deps{
		clients:  clientsFromConfig(defaultConfig()),
		sessions: NewSessionStore(),
	}

	body := `{"title": "test", "model": "gpt-4"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/session", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	d.handleCreateSession(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result api.CreateSessionResult
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Session.Title != "test" {
		t.Errorf("expected title 'test', got %s", result.Session.Title)
	}
}

func TestHandleCreateSession_InvalidJSON(t *testing.T) {
	d := &deps{
		clients:  clientsFromConfig(defaultConfig()),
		sessions: NewSessionStore(),
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/session", bytes.NewBufferString(`not json`))
	rr := httptest.NewRecorder()

	d.handleCreateSession(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleSendMessage_MissingID(t *testing.T) {
	d := &deps{
		clients:  clientsFromConfig(defaultConfig()),
		sessions: NewSessionStore(),
	}

	body := `{"id": "m1", "content": "hello"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/session//message", bytes.NewBufferString(body))
	req.SetPathValue("id", "")
	rr := httptest.NewRecorder()

	d.handleSendMessage(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleSendMessage_InvalidJSON(t *testing.T) {
	d := &deps{
		clients:  clientsFromConfig(defaultConfig()),
		sessions: NewSessionStore(),
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/session/s1/message", bytes.NewBufferString(`not json`))
	req.SetPathValue("id", "s1")
	rr := httptest.NewRecorder()

	d.handleSendMessage(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleTerminal_ValidJSON(t *testing.T) {
	body := `{"command": "ls", "cols": 80, "rows": 24}`
	req := httptest.NewRequest(http.MethodPost, "/v1/terminal", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	handleTerminal(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

func TestHandleTerminalInput_ValidJSON(t *testing.T) {
	body := `{"data": "ls\n"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/terminal/t1/input", bytes.NewBufferString(body))
	req.SetPathValue("id", "t1")
	rr := httptest.NewRecorder()

	handleTerminalInput(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

func TestHandleTerminalInput_MissingID(t *testing.T) {
	body := `{"data": "ls\n"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/terminal//input", bytes.NewBufferString(body))
	req.SetPathValue("id", "")
	rr := httptest.NewRecorder()

	handleTerminalInput(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleTerminalStream_MissingID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/terminal//stream", nil)
	req.SetPathValue("id", "")
	rr := httptest.NewRecorder()

	handleTerminalStream(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleInputMouse_ValidJSON(t *testing.T) {
	body := `{"x": 100, "y": 200, "action": "click"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/input/mouse", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	handleInputMouse(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

func TestHandleInputKey_ValidJSON(t *testing.T) {
	body := `{"key": "enter", "action": "press"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/input/key", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	handleInputKey(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

func TestHandleCapture_ValidJSON(t *testing.T) {
	body := `{"fullscreen": true, "attach": false}`
	req := httptest.NewRequest(http.MethodPost, "/v1/screen/capture", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	handleCapture(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

func TestHandleVoiceTranscribe_ValidJSON(t *testing.T) {
	body := `{"audio_id": "a1", "lang": "ru"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/voice/transcribe", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	handleVoiceTranscribe(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

func TestHandleDeviceInfo(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/device/info", nil)
	rr := httptest.NewRecorder()

	handleDeviceInfo(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

func TestHandleEvents(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/events", nil)
	rr := httptest.NewRecorder()

	handleEvents(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected text/event-stream, got %s", ct)
	}
}

func TestHandleUploadAttachment(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/attachment", nil)
	rr := httptest.NewRecorder()

	handleUploadAttachment(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 501, got %d", rr.Code)
	}
}

func TestNewRouter_AuthRequired(t *testing.T) {
	d := &deps{
		clients:  clientsFromConfig(defaultConfig()),
		sessions: NewSessionStore(),
	}
	router := newRouter(d)

	req := httptest.NewRequest(http.MethodGet, "/v1/status", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", rr.Code)
	}
}

func TestNewRouter_AuthBypass(t *testing.T) {
	d := &deps{
		clients:  clientsFromConfig(defaultConfig()),
		sessions: NewSessionStore(),
	}
	router := newRouter(d)

	req := httptest.NewRequest(http.MethodGet, "/v1/status", nil)
	req.Header.Set("Authorization", "Bearer dev-token")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 with valid token, got %d", rr.Code)
	}
}

func TestNewRouter_SessionCreate(t *testing.T) {
	d := &deps{
		clients:  clientsFromConfig(defaultConfig()),
		sessions: NewSessionStore(),
	}
	router := newRouter(d)

	body := `{"title": "test session"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/session", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer dev-token")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result api.CreateSessionResult
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Session.Title != "test session" {
		t.Errorf("expected title 'test session', got %s", result.Session.Title)
	}
}
