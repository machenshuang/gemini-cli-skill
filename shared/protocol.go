package shared

import (
	"errors"
	"time"
)

var ErrConcurrentLimit = errors.New("concurrent limit reached")

// Backend type
type Backend string

const (
	BackendGemini Backend = "gemini"
	BackendKimi   Backend = "kimi"
)

// Task state machine
type TaskState string

const (
	TaskStatePending   TaskState = "pending"
	TaskStateRunning   TaskState = "running"
	TaskStateCompleted TaskState = "completed"
	TaskStateFailed    TaskState = "failed"
	TaskStateStopped   TaskState = "stopped"
	TaskStateTimeout   TaskState = "timeout"
)

// Approval mode
type ApprovalMode string

const (
	ApprovalModeDefault  ApprovalMode = "default"
	ApprovalModeAutoEdit ApprovalMode = "auto_edit"
	ApprovalModeYolo     ApprovalMode = "yolo"
)

// Verbosity level
type Verbosity string

const (
	VerbosityMinimal Verbosity = "minimal"
	VerbosityNormal  Verbosity = "normal"
	VerbosityFull    Verbosity = "full"
)

// CLI stream-json events (compatible with both Gemini and Kimi)

type CliEventType string

const (
	CliEventTypeInit       CliEventType = "init"
	CliEventTypeMessage    CliEventType = "message"
	CliEventTypeToolUse    CliEventType = "tool_use"
	CliEventTypeToolResult CliEventType = "tool_result"
	CliEventTypeResult     CliEventType = "result"
)

type CliInitEvent struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id"`
	Model     string `json:"model"`
}

type CliMessageEvent struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp,omitempty"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Delta     bool   `json:"delta,omitempty"`
}

type CliToolUseEvent struct {
	Type       string                 `json:"type"`
	Timestamp  string                 `json:"timestamp,omitempty"`
	ToolName   string                 `json:"tool_name"`
	ToolID     string                 `json:"tool_id"`
	Parameters map[string]interface{} `json:"parameters"`
}

type CliToolResultEvent struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp,omitempty"`
	ToolID    string `json:"tool_id"`
	Status    string `json:"status"`
	Output    string `json:"output"`
}

type TaskStats struct {
	TotalTokens  int `json:"total_tokens,omitempty"`
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
	DurationMs   int `json:"duration_ms,omitempty"`
	ToolCalls    int `json:"tool_calls,omitempty"`
}

type CliResultEvent struct {
	Type      string    `json:"type"`
	Timestamp string    `json:"timestamp,omitempty"`
	Status    string    `json:"status"`
	Stats     TaskStats `json:"stats"`
}

// Task data structures

type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ToolCall struct {
	Name       string                 `json:"name"`
	ToolID     string                 `json:"tool_id"`
	Status     string                 `json:"status"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Output     string                 `json:"output,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

type TaskResult struct {
	FinalResponse string    `json:"finalResponse"`
	Stats         TaskStats `json:"stats"`
}

type TaskRecord struct {
	ID           string         `json:"id"`
	PID          int            `json:"pid,omitempty"`
	State        TaskState      `json:"state"`
	Prompt       string         `json:"prompt"`
	WorkingDir   string         `json:"workingDir"`
	Model        string         `json:"model,omitempty"`
	ApprovalMode ApprovalMode   `json:"approvalMode"`
	Timeout      int            `json:"timeout"`
	Tags         []string       `json:"tags"`
	Backend      Backend        `json:"backend"`
	Thinking     bool           `json:"thinking,omitempty"`
	SessionID    string         `json:"sessionId,omitempty"`
	Messages     []Message      `json:"messages"`
	ToolCalls    []ToolCall     `json:"toolCalls"`
	Result       *TaskResult    `json:"result,omitempty"`
	Error        string         `json:"error,omitempty"`
	StartedAt    time.Time      `json:"startedAt"`
	CompletedAt  *time.Time     `json:"completedAt,omitempty"`
}

type TaskSummary struct {
	ID            string    `json:"id"`
	State         TaskState `json:"state"`
	PromptPreview string    `json:"promptPreview"`
	StartedAt     time.Time `json:"startedAt"`
	ElapsedMs     int64     `json:"elapsedMs"`
	Tags          []string  `json:"tags"`
}

// Daemon RPC protocol

type RpcAction string

const (
	RpcActionStart    RpcAction = "start"
	RpcActionStatus   RpcAction = "status"
	RpcActionStop     RpcAction = "stop"
	RpcActionList     RpcAction = "list"
	RpcActionShutdown RpcAction = "shutdown"
)

type RpcRequest struct {
	Action RpcAction              `json:"action"`
	Params map[string]interface{} `json:"params"`
}

type RpcResponse struct {
	Ok    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// Start params

type StartParams struct {
	Prompt       string       `json:"prompt"`
	WorkingDir   string       `json:"workingDir,omitempty"`
	Model        string       `json:"model,omitempty"`
	ApprovalMode ApprovalMode `json:"approvalMode,omitempty"`
	Timeout      int          `json:"timeout,omitempty"`
	Tags         []string     `json:"tags,omitempty"`
	Backend      Backend      `json:"backend,omitempty"`
	Thinking     bool         `json:"thinking,omitempty"`
}

// Status params

type StatusParams struct {
	TaskID    string    `json:"taskId"`
	Verbosity Verbosity `json:"verbosity,omitempty"`
	Tail      int       `json:"tail,omitempty"`
}

// Stop params

type StopParams struct {
	TaskID string `json:"taskId"`
	Force  bool   `json:"force,omitempty"`
}

// List params

type ListParams struct {
	State []TaskState `json:"state,omitempty"`
	Tags  []string    `json:"tags,omitempty"`
	Limit int         `json:"limit,omitempty"`
}

// Config

type RunnerConfig struct {
	MaxConcurrent       int          `json:"maxConcurrent"`
	DefaultTimeout      int          `json:"defaultTimeout"`
	DefaultApprovalMode ApprovalMode `json:"defaultApprovalMode"`
	DefaultBackend      Backend      `json:"defaultBackend"`
	DefaultThinking     bool         `json:"defaultThinking,omitempty"`
}
