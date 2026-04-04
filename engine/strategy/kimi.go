package strategy

import (
	"bufio"
	"cli-agent-go/shared"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// KimiEvent Kimi 事件类型
type KimiEvent struct {
	Role        string `json:"role,omitempty"`
	Content     interface{} `json:"content,omitempty"`
	ToolCalls   []KimiToolCall `json:"tool_calls,omitempty"`
	ToolCallID  string `json:"tool_call_id,omitempty"`
	Error       string `json:"error,omitempty"`
}

type KimiToolCall struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// KimiOutputHandler Kimi 输出处理器
type KimiOutputHandler struct {
	emitter EventEmitter
	started bool
	sid     string
}

// NewKimiOutputHandler 创建 Kimi 输出处理器
func NewKimiOutputHandler(emitter EventEmitter) *KimiOutputHandler {
	return &KimiOutputHandler{
		emitter: emitter,
		started: false,
		sid:     "",
	}
}

// HandleOutput 处理输出流
func (h *KimiOutputHandler) HandleOutput(stdout interface{}) {
	reader, ok := stdout.(io.Reader)
	if !ok {
		return
	}
	
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var data KimiEvent
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			// non-json line, skip
			continue
		}
		h.parseEvent(&data)
	}
}

// Destroy 清理资源
func (h *KimiOutputHandler) Destroy() {
	// Nothing to clean up
}

func (h *KimiOutputHandler) parseEvent(data *KimiEvent) {
	// Initialize on first event
	if !h.started {
		h.started = true
		h.sid = fmt.Sprintf("kimi-%d", time.Now().UnixMilli())

		initEvt := shared.CliInitEvent{
			Type:      "init",
			Timestamp: time.Now().Format(time.RFC3339),
			SessionID: h.sid,
			Model:     "kimi-default",
		}
		h.emitter.Emit("init", initEvt)
	}

	// Handle error
	if data.Error != "" {
		h.emitter.Emit("error", fmt.Errorf(data.Error))
		return
	}

	// Handle assistant message with tool_calls
	if data.Role == "assistant" && len(data.ToolCalls) > 0 {
		for _, tc := range data.ToolCalls {
			var params map[string]interface{}
			json.Unmarshal([]byte(tc.Function.Arguments), &params)
			
			toolEvt := shared.CliToolUseEvent{
				Type:       "tool_use",
				Timestamp:  time.Now().Format(time.RFC3339),
				ToolName:   tc.Function.Name,
				ToolID:     tc.ID,
				Parameters: params,
			}
			h.emitter.Emit("tool_use", toolEvt)
		}
		return
	}

	// Handle tool result (role: tool)
	if data.Role == "tool" && data.ToolCallID != "" {
		content := h.extractContent(data)
		toolResultEvt := shared.CliToolResultEvent{
			Type:      "tool_result",
			Timestamp: time.Now().Format(time.RFC3339),
			ToolID:    data.ToolCallID,
			Status:    "success",
			Output:    content,
		}
		h.emitter.Emit("tool_result", toolResultEvt)
		return
	}

	// Handle final assistant message (no tool_calls)
	if data.Role == "assistant" && data.Content != nil {
		content := h.extractContent(data)

		if content != "" {
			msgEvt := shared.CliMessageEvent{
				Type:      "message",
				Timestamp: time.Now().Format(time.RFC3339),
				Role:      "assistant",
				Content:   content,
			}
			h.emitter.Emit("message", msgEvt)
		}

		// Emit result event for final message
		stats := shared.TaskStats{}
		resultEvt := shared.CliResultEvent{
			Type:      "result",
			Timestamp: time.Now().Format(time.RFC3339),
			Status:    "success",
			Stats:     stats,
		}
		h.emitter.Emit("result", resultEvt)
	}
}

func (h *KimiOutputHandler) extractContent(data *KimiEvent) string {
	if data.Content == nil {
		return ""
	}

	// Handle string content
	if s, ok := data.Content.(string); ok {
		return strings.TrimSpace(s)
	}

	// Handle array content
	arr, ok := data.Content.([]interface{})
	if !ok {
		return ""
	}

	var result strings.Builder
	for _, item := range arr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		// Skip thinking content, only keep text
		if obj["type"] == "text" {
			if text, ok := obj["text"].(string); ok {
				result.WriteString(text)
			}
		}
	}
	return strings.TrimSpace(result.String())
}

// KimiStrategy Kimi CLI 策略
type KimiStrategy struct {
	name string
}

// NewKimiStrategy 创建 Kimi 策略
func NewKimiStrategy() *KimiStrategy {
	return &KimiStrategy{name: "kimi"}
}

// Name 返回策略名称
func (s *KimiStrategy) Name() string {
	return s.name
}

// BuildCommand 构建命令
func (s *KimiStrategy) BuildCommand(opts ExecutorOpts) (cmd string, args []string, useStdin bool) {
	args = []string{"--print", "--output-format", "stream-json"}

	if opts.Model != "" {
		args = append(args, "-m", opts.Model)
	}

	// Kimi uses -y or --yolo for auto-approve
	if opts.ApprovalMode == shared.ApprovalModeYolo {
		args = append(args, "-y")
	}

	// Thinking mode
	if opts.Thinking {
		args = append(args, "--thinking")
	}

	// Kimi uses -p for prompt (not stdin)
	args = append(args, "-p", opts.Prompt)

	return "kimi", args, false
}

// CreateOutputHandler 创建输出处理器
func (s *KimiStrategy) CreateOutputHandler(emitter EventEmitter) OutputHandler {
	return NewKimiOutputHandler(emitter)
}
