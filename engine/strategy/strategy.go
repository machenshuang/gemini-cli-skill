package strategy

import (
	"cli-agent-go/shared"
)

// ExecutorOpts 执行器选项
type ExecutorOpts struct {
	Prompt       string
	WorkingDir   string
	Model        string
	ApprovalMode shared.ApprovalMode
	Timeout      int
	Backend      shared.Backend
	Thinking     bool
}

// OutputHandler 输出处理器接口
type OutputHandler interface {
	// HandleOutput 处理输出流
	HandleOutput(stdout interface{})
	// Destroy 清理资源
	Destroy()
}

// EventEmitter 事件发射器接口
type EventEmitter interface {
	Emit(event string, data interface{})
	On(event string, handler func(interface{}))
}

// CliStrategy CLI 策略接口
// 每个后端（Gemini、Kimi 等）实现此接口
type CliStrategy interface {
	// Name 策略名称
	Name() string
	// BuildCommand 构建命令参数
	BuildCommand(opts ExecutorOpts) (cmd string, args []string, useStdin bool)
	// CreateOutputHandler 创建输出处理器
	CreateOutputHandler(emitter EventEmitter) OutputHandler
}

// StrategyFactory 创建策略工厂
func CreateStrategy(backend shared.Backend) CliStrategy {
	switch backend {
	case shared.BackendGemini:
		return NewGeminiStrategy()
	case shared.BackendKimi:
		return NewKimiStrategy()
	default:
		return NewKimiStrategy()
	}
}
