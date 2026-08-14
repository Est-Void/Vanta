package api

import "time"

type Code string

const (
	ErrBadRequest   Code = "bad_request"
	ErrUnauthorized Code = "unauthorized"
	ErrNotFound     Code = "not_found"
	ErrDenied       Code = "permission_denied"
	ErrConflict     Code = "conflict"
	ErrInternal     Code = "internal"
)

func StatusCode(c Code) int {
	switch c {
	case ErrBadRequest:
		return 400
	case ErrUnauthorized:
		return 401
	case ErrDenied:
		return 403
	case ErrNotFound:
		return 404
	case ErrConflict:
		return 409
	case ErrInternal:
		return 500
	default:
		return 500
	}
}

type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details"`
}

type Role string

const (
	RoleUser   Role = "user"
	RoleSystem Role = "system"
	RoleAgent  Role = "agent"
)

type AuthScope string

const (
	ScopeTerminal AuthScope = "terminal"
	ScopeDevice   AuthScope = "device"
	ScopeInput    AuthScope = "input"
	ScopeScreen   AuthScope = "screen"
	ScopeVoice    AuthScope = "voice"
	ScopeAgent    AuthScope = "agent"
)

type AttInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	MIME string `json:"mime"`
	Size int64  `json:"size"`
}

type Session struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	CreatedAt   time.Time `json:"created_at"`
	LastMessage time.Time `json:"last_message"`
	Attachments []AttInfo `json:"attachments"`
}

type Message struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	Role        Role      `json:"role"`
	Content     string    `json:"content"`
	Attachments []AttInfo `json:"attachments"`
	CreatedAt   time.Time `json:"created_at"`
}

type Status struct {
	SessionID string         `json:"session_id"`
	Data      map[string]any `json:"data"`
}
