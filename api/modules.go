package api

type CreateSessionRequest struct {
	Title string `json:"title,omitempty"`
	Model string `json:"model,omitempty"`
}

type CreateSessionResult struct {
	Session `json:"session"`
}

type ListSessionsResult struct {
	Sessions []Session `json:"sessions"`
}

type SendMessage struct {
	ID          string   `json:"id"`
	Content     string   `json:"content"`
	Attachments []string `json:"attachments,omitempty"`
	Model       string   `json:"model,omitempty"`
	Streaming   bool     `json:"streaming,omitempty"`
	ResumeID    string   `json:"resume_id,omitempty"`
}

type CreateTerminalRequest struct {
	Command string   `json:"command,omitempty"`
	CWD     string   `json:"cwd,omitempty"`
	Env     []string `json:"env,omitempty"`
	Cols    uint16   `json:"cols"`
	Rows    uint16   `json:"rows"`
}

type TerminalSession struct {
	ID       string `json:"id"`
	CWD      string `json:"cwd"`
	Command  string `json:"command"`
	PID      int    `json:"pid"`
	ExitCode *int   `json:"exit_code,omitempty"`
}

type WriteInput struct {
	Data string `json:"data"`
}

type TerminalResize struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

type CaptureRequest struct {
	FullScreen bool `json:"fullscreen"`
	Monitor    int  `json:"monitor,omitempty"`
	Attach     bool `json:"attach"`
}

type CaptureResult struct {
	AttID  string `json:"att_id,omitempty"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type MouseButton string

const (
	MouseLeft   MouseButton = "left"
	MouseRight  MouseButton = "right"
	MouseMiddle MouseButton = "middle"
	MouseScroll MouseButton = "scroll"
)

type MouseEvent struct {
	X      int         `json:"x"`
	Y      int         `json:"y"`
	Rel    bool        `json:"rel,omitempty"`
	Button MouseButton `json:"button,omitempty"`
	Action string      `json:"action"`
	Amount int         `json:"amount,omitempty"`
}

type KeyEvent struct {
	Key    string `json:"key"`
	Action string `json:"action"`
	Text   string `json:"text,omitempty"`
}

type CPUInfo struct {
	Model string  `json:"model"`
	Cores int     `json:"cores"`
	Load  float64 `json:"load"`
}

type DeviceInfo struct {
	Hostname string  `json:"hostname"`
	OS       string  `json:"os"`
	Kernel   string  `json:"kernel"`
	Arch     string  `json:"arch"`
	CPU      CPUInfo `json:"cpu"`
	MemUsed  uint64  `json:"mem_used"`
	MemTotal uint64  `json:"mem_total"`
	Uptime   int64   `json:"uptime_seconds"`
}

type TranscribeRequest struct {
	AudioID string `json:"audio_id"`
	Lang    string `json:"lang,omitempty"`
}

type TranscribeResult struct {
	Text string `json:"text"`
}
