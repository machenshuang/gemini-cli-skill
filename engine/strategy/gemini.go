package strategy

import (
	"bufio"
	"cli-agent-go/shared"
	"encoding/json"
	"io"
	"strings"
)

// GeminiOutputHandler Gemini 输出处理器
type GeminiOutputHandler struct {
	emitter EventEmitter
}

// NewGeminiOutputHandler 创建 Gemini 输出处理器
func NewGeminiOutputHandler(emitter EventEmitter) *GeminiOutputHandler {
	return &GeminiOutputHandler{
		emitter: emitter,
	}
}

// HandleOutput 处理输出流
func (h *GeminiOutputHandler) HandleOutput(stdout interface{}) {
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

		var evt map[string]interface{}
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			// non-json line, skip
			continue
		}
		h.dispatch(evt)
	}
}

// Destroy 清理资源
func (h *GeminiOutputHandler) Destroy() {
	// Nothing to clean up
}

func (h *GeminiOutputHandler) dispatch(evt map[string]interface{}) {
	evtType, ok := evt["type"].(string)
	if !ok {
		return
	}

	switch evtType {
	case "init":
		initEvt := shared.CliInitEvent{
			Type:      "init",
			Timestamp: getString(evt, "timestamp"),
			SessionID: getString(evt, "session_id"),
			Model:     getString(evt, "model"),
		}
		h.emitter.Emit("init", initEvt)

	case "message":
		msgEvt := shared.CliMessageEvent{
			Type:      "message",
			Timestamp: getString(evt, "timestamp"),
			Role:      getString(evt, "role"),
			Content:   getString(evt, "content"),
			Delta:     getBool(evt, "delta"),
		}
		h.emitter.Emit("message", msgEvt)

	case "tool_use":
		toolEvt := shared.CliToolUseEvent{
			Type:       "tool_use",
			Timestamp:  getString(evt, "timestamp"),
			ToolName:   getString(evt, "tool_name"),
			ToolID:     getString(evt, "tool_id"),
			Parameters: getMap(evt, "parameters"),
		}
		h.emitter.Emit("tool_use", toolEvt)

	case "tool_result":
		toolResultEvt := shared.CliToolResultEvent{
			Type:      "tool_result",
			Timestamp: getString(evt, "timestamp"),
			ToolID:    getString(evt, "tool_id"),
			Status:    getString(evt, "status"),
			Output:    getString(evt, "output"),
		}
		h.emitter.Emit("tool_result", toolResultEvt)

	case "result":
		resultEvt := shared.CliResultEvent{
			Type:      "result",
			Timestamp: getString(evt, "timestamp"),
			Status:    getString(evt, "status"),
			Stats:     getStats(evt, "stats"),
		}
		h.emitter.Emit("result", resultEvt)
	}
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key].(map[string]interface{}); ok {
		return v
	}
	return nil
}

func getStats(m map[string]interface{}, key string) shared.TaskStats {
	stats := shared.TaskStats{}
	if v, ok := m[key].(map[string]interface{}); ok {
		if n, ok := v["total_tokens"].(float64); ok {
			stats.TotalTokens = int(n)
		}
		if n, ok := v["input_tokens"].(float64); ok {
			stats.InputTokens = int(n)
		}
		if n, ok := v["output_tokens"].(float64); ok {
			stats.OutputTokens = int(n)
		}
		if n, ok := v["duration_ms"].(float64); ok {
			stats.DurationMs = int(n)
		}
		if n, ok := v["tool_calls"].(float64); ok {
			stats.ToolCalls = int(n)
		}
	}
	return stats
}

// GeminiStrategy Gemini CLI 策略
type GeminiStrategy struct {
	name string
}

// NewGeminiStrategy 创建 Gemini 策略
func NewGeminiStrategy() *GeminiStrategy {
	return &GeminiStrategy{name: "gemini"}
}

// Name 返回策略名称
func (s *GeminiStrategy) Name() string {
	return s.name
}

// BuildCommand 构建命令
func (s *GeminiStrategy) BuildCommand(opts ExecutorOpts) (cmd string, args []string, useStdin bool) {
	args = []string{"--output-format", "stream-json"}

	if opts.Model != "" {
		args = append(args, "-m", opts.Model)
	}

	if opts.ApprovalMode == shared.ApprovalModeYolo {
		args = append(args, "-y")
	} else if opts.ApprovalMode != "" {
		// 有明确的 approval mode 才传，否则让 gemini CLI 使用自身默认值
		args = append(args, "--approval-mode", string(opts.ApprovalMode))
	}

	return "gemini", args, true
}

// CreateOutputHandler 创建输出处理器
func (s *GeminiStrategy) CreateOutputHandler(emitter EventEmitter) OutputHandler {
	return NewGeminiOutputHandler(emitter)
}
