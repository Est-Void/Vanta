package api

import "time"

type EventType string

const (
	EvToken          EventType = "token"
	EvMessage        EventType = "message"
	EvToolCall       EventType = "tool_call"
	EvToolResult     EventType = "tool_result"
	EvApproval       EventType = "approval"
	EvTerminalOutput EventType = "terminal_output"
	EvTerminalExit   EventType = "terminal_exit"
	EvError          EventType = "error"
	EvStopped        EventType = "stopped"
	EvDone           EventType = "done"
)

type Event struct {
	Type      EventType `json:"type"`
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Data      any       `json:"data"`
}

type TokenData struct {
	Text string `json:"text"`
}

type MessageData struct {
	Message `json:"message"`
}

type ToolStatus string

const (
	ToolRunning ToolStatus = "running"
	ToolDone    ToolStatus = "done"
	ToolFailed  ToolStatus = "failed"
)

type ToolCallData struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Status ToolStatus     `json:"status"`
	Args   map[string]any `json:"args,omitempty"`
}

type ToolResultData struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

type ApprovalData struct {
	Approval `json:"approval"`
}

type Approval struct {
	ID        string    `json:"id"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Reason    string    `json:"reason,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
}

type TerminalOutputData struct {
	Data string `json:"data"`
	More bool   `json:"more"`
}

type TerminalExitData struct {
	ExitCode int `json:"exit_code"`
}

type ErrorData struct {
	Error `json:"error"`
}

type TokenUsage struct {
	Prompt     int `json:"prompt"`
	Completion int `json:"completion"`
}

type DoneData struct {
	MessageID string        `json:"message_id,omitempty"`
	Duration  time.Duration `json:"duration"`
	Usage     *TokenUsage   `json:"usage,omitempty"`
	Err       *Error        `json:"error,omitempty"`
}
