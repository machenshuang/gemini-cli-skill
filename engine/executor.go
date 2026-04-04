package engine

import (
	"cli-agent-go/engine/strategy"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// ExecutorEvent 执行器事件
type ExecutorEvent struct {
	Type string
	Data interface{}
}

// SimpleEventEmitter 简单的事件发射器实现
type SimpleEventEmitter struct {
	handlers map[string][]func(interface{})
	mu       sync.RWMutex
}

// NewSimpleEventEmitter 创建事件发射器
func NewSimpleEventEmitter() *SimpleEventEmitter {
	return &SimpleEventEmitter{
		handlers: make(map[string][]func(interface{})),
	}
}

// Emit 发射事件
func (e *SimpleEventEmitter) Emit(event string, data interface{}) {
	e.mu.RLock()
	handlers := e.handlers[event]
	e.mu.RUnlock()
	for _, h := range handlers {
		go h(data)
	}
}

// On 注册事件处理器
func (e *SimpleEventEmitter) On(event string, handler func(interface{})) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[event] = append(e.handlers[event], handler)
}

// CliExecutor 包装单个 CLI 进程
type CliExecutor struct {
	opts       strategy.ExecutorOpts
	strategy   strategy.CliStrategy
	handler    strategy.OutputHandler
	emitter    *SimpleEventEmitter
	proc       *exec.Cmd
	cancelFunc context.CancelFunc
	alive      bool
	mu         sync.Mutex
	timer      *time.Timer
}

// NewCliExecutor 创建 CLI 执行器
func NewCliExecutor(opts strategy.ExecutorOpts) *CliExecutor {
	s := strategy.CreateStrategy(opts.Backend)
	emitter := NewSimpleEventEmitter()
	handler := s.CreateOutputHandler(emitter)

	return &CliExecutor{
		opts:     opts,
		strategy: s,
		handler:  handler,
		emitter:  emitter,
		alive:    false,
	}
}

// SessionID 获取会话 ID
func (e *CliExecutor) SessionID() string {
	return ""
}

// PID 获取进程 ID
func (e *CliExecutor) PID() int {
	if e.proc != nil && e.proc.Process != nil {
		return e.proc.Process.Pid
	}
	return 0
}

// IsAlive 检查是否存活
func (e *CliExecutor) IsAlive() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.alive
}

// On 注册事件处理器
func (e *CliExecutor) On(event string, handler func(interface{})) {
	e.emitter.On(event, handler)
}

// Launch 启动执行器
func (e *CliExecutor) Launch() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.alive {
		return fmt.Errorf("executor already running")
	}
	e.alive = true

	cmdStr, args, useStdin := e.strategy.BuildCommand(e.opts)

	ctx, cancel := context.WithCancel(context.Background())
	e.cancelFunc = cancel

	e.proc = exec.CommandContext(ctx, cmdStr, args...)
	e.proc.Dir = e.opts.WorkingDir
	e.proc.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Feed prompt via stdin if needed
	if useStdin {
		e.proc.Stdin = stringsToReader(e.opts.Prompt)
	}

	// Get stdout pipe
	stdout, err := e.proc.StdoutPipe()
	if err != nil {
		e.alive = false
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Get stderr pipe
	stderr, err := e.proc.StderrPipe()
	if err != nil {
		e.alive = false
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := e.proc.Start(); err != nil {
		e.alive = false
		return fmt.Errorf("failed to start process: %w", err)
	}

	// Handle output using strategy
	go e.handler.HandleOutput(stdout)
	
	// Handle stderr
	go io.Copy(io.Discard, stderr)

	// Wait for process to exit
	go func() {
		err := e.proc.Wait()
		e.teardown()
		if err != nil {
			e.emitter.Emit("exit", 1)
		} else {
			e.emitter.Emit("exit", 0)
		}
	}()

	// Arm timeout
	if e.opts.Timeout > 0 {
		e.timer = time.AfterFunc(time.Duration(e.opts.Timeout)*time.Second, func() {
			e.emitter.Emit("timeout", nil)
			e.Kill(false)
		})
	}

	return nil
}

// Kill 杀死进程
func (e *CliExecutor) Kill(force bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.proc == nil || e.proc.Process == nil || !e.alive {
		return
	}

	sig := syscall.SIGTERM
	if force {
		sig = syscall.SIGKILL
	}

	// Try to kill the process group
	syscall.Kill(-e.proc.Process.Pid, sig)

	// If not force, escalate after 5s
	if !force {
		time.AfterFunc(5*time.Second, func() {
			e.mu.Lock()
			defer e.mu.Unlock()
			if e.proc != nil && e.proc.Process != nil && e.alive {
				syscall.Kill(-e.proc.Process.Pid, syscall.SIGKILL)
			}
		})
	}
}

func (e *CliExecutor) teardown() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.alive = false
	e.handler.Destroy()
	if e.timer != nil {
		e.timer.Stop()
		e.timer = nil
	}
	if e.cancelFunc != nil {
		e.cancelFunc()
	}
	e.proc = nil
}

func stringsToReader(s string) io.Reader {
	return &stringReader{s: s, i: 0}
}

type stringReader struct {
	s string
	i int
}

func (r *stringReader) Read(p []byte) (n int, err error) {
	if r.i >= len(r.s) {
		return 0, io.EOF
	}
	n = copy(p, r.s[r.i:])
	r.i += n
	return n, nil
}

// Legacy alias
var GeminiExecutor = NewCliExecutor

// MockOsStdin 用于测试
var MockOsStdin = os.Stdin
