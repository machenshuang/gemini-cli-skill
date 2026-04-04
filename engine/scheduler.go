package engine

import (
	"cli-agent-go/engine/strategy"
	"cli-agent-go/shared"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Scheduler 内存任务调度器
type Scheduler struct {
	tasks       map[string]*shared.TaskRecord
	executors   map[string]*CliExecutor
	deltaBuffers map[string]string
	config      shared.RunnerConfig
	sweepTimer  *time.Timer
	mu          sync.RWMutex
}

// NewScheduler 创建调度器
func NewScheduler(config shared.RunnerConfig) *Scheduler {
	s := &Scheduler{
		tasks:        make(map[string]*shared.TaskRecord),
		executors:    make(map[string]*CliExecutor),
		deltaBuffers: make(map[string]string),
		config:       config,
	}

	s.sweepTimer = time.AfterFunc(shared.CLEANUP_INTERVAL_MS*time.Millisecond, func() {
		s.sweep()
	})

	return s
}

// CreateTask 创建任务
func (s *Scheduler) CreateTask(params shared.StartParams) (*shared.TaskRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	running := s.countByState(shared.TaskStateRunning)
	if running >= s.config.MaxConcurrent {
		return nil, shared.ErrConcurrentLimit
	}

	id := generateID()
	now := time.Now()
	backend := params.Backend
	if backend == "" {
		backend = s.config.DefaultBackend
	}

	approvalMode := params.ApprovalMode
	if approvalMode == "" {
		approvalMode = s.config.DefaultApprovalMode
	}

	timeout := params.Timeout
	if timeout == 0 {
		timeout = s.config.DefaultTimeout
	}

	task := &shared.TaskRecord{
		ID:           id,
		State:        shared.TaskStateRunning,
		Prompt:       params.Prompt,
		WorkingDir:   params.WorkingDir,
		Model:        params.Model,
		ApprovalMode: approvalMode,
		Timeout:      timeout,
		Tags:         params.Tags,
		Backend:      backend,
		Thinking:     params.Thinking,
		Messages:     []shared.Message{},
		ToolCalls:    []shared.ToolCall{},
		StartedAt:    now,
	}

	s.tasks[id] = task

	exec := NewCliExecutor(strategy.ExecutorOpts{
		Prompt:       task.Prompt,
		WorkingDir:   task.WorkingDir,
		Model:        task.Model,
		ApprovalMode: task.ApprovalMode,
		Timeout:      task.Timeout,
		Backend:      backend,
		Thinking:     task.Thinking,
	})

	s.executors[id] = exec
	s.deltaBuffers[id] = ""
	s.wireEvents(id, exec, task)

	if err := exec.Launch(); err != nil {
		task.State = shared.TaskStateFailed
		task.Error = err.Error()
		completedAt := time.Now()
		task.CompletedAt = &completedAt
		return task, nil
	}

	task.PID = exec.PID()
	return task, nil
}

// GetTask 获取任务
func (s *Scheduler) GetTask(id string) *shared.TaskRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tasks[id]
}

// StopTask 停止任务
func (s *Scheduler) StopTask(id string, force bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := s.tasks[id]
	exec := s.executors[id]
	if task == nil || exec == nil || task.State != shared.TaskStateRunning {
		return false
	}

	exec.Kill(force)
	task.State = shared.TaskStateStopped
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	return true
}

// ListTasks 列出任务
func (s *Scheduler) ListTasks(filter *shared.ListParams) []shared.TaskSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var items []*shared.TaskRecord
	for _, t := range s.tasks {
		items = append(items, t)
	}

	if filter != nil {
		if len(filter.State) > 0 {
			var filtered []*shared.TaskRecord
			for _, t := range items {
				for _, state := range filter.State {
					if t.State == state {
						filtered = append(filtered, t)
						break
					}
				}
			}
			items = filtered
		}

		if len(filter.Tags) > 0 {
			var filtered []*shared.TaskRecord
			for _, t := range items {
				for _, tag := range filter.Tags {
					if contains(t.Tags, tag) {
						filtered = append(filtered, t)
						break
					}
				}
			}
			items = filtered
		}
	}

	// Sort by startedAt desc
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].StartedAt.Before(items[j].StartedAt) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	limit := 20
	if filter != nil && filter.Limit > 0 {
		limit = filter.Limit
	}
	if len(items) > limit {
		items = items[:limit]
	}

	var summaries []shared.TaskSummary
	for _, t := range items {
		preview := t.Prompt
		if len(preview) > 60 {
			preview = preview[:60] + "..."
		}
		elapsed := time.Since(t.StartedAt).Milliseconds()
		if t.CompletedAt != nil {
			elapsed = t.CompletedAt.Sub(t.StartedAt).Milliseconds()
		}
		summaries = append(summaries, shared.TaskSummary{
			ID:            t.ID,
			State:         t.State,
			PromptPreview: preview,
			StartedAt:     t.StartedAt,
			ElapsedMs:     elapsed,
			Tags:          t.Tags,
		})
	}

	return summaries
}

// RunningCount 获取运行中任务数
func (s *Scheduler) RunningCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.countByState(shared.TaskStateRunning)
}

// TotalCount 获取总任务数
func (s *Scheduler) TotalCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tasks)
}

// Shutdown 关闭调度器
func (s *Scheduler) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sweepTimer != nil {
		s.sweepTimer.Stop()
		s.sweepTimer = nil
	}
	for _, exec := range s.executors {
		exec.Kill(true)
	}
	s.executors = make(map[string]*CliExecutor)
	s.deltaBuffers = make(map[string]string)
}

func (s *Scheduler) countByState(state shared.TaskState) int {
	n := 0
	for _, t := range s.tasks {
		if t.State == state {
			n++
		}
	}
	return n
}

func (s *Scheduler) wireEvents(id string, exec *CliExecutor, task *shared.TaskRecord) {
	exec.On("init", func(data interface{}) {
		if evt, ok := data.(shared.CliInitEvent); ok {
			s.mu.Lock()
			task.SessionID = evt.SessionID
			s.mu.Unlock()
		}
	})

	exec.On("message", func(data interface{}) {
		if evt, ok := data.(shared.CliMessageEvent); ok {
			s.mu.Lock()
			defer s.mu.Unlock()
			if evt.Role == "assistant" && evt.Delta {
				buf := s.deltaBuffers[id]
				s.deltaBuffers[id] = buf + evt.Content
			} else {
				task.Messages = append(task.Messages, shared.Message{
					Role:      evt.Role,
					Content:   evt.Content,
					Timestamp: time.Now(),
				})
			}
		}
	})

	exec.On("tool_use", func(data interface{}) {
		if evt, ok := data.(shared.CliToolUseEvent); ok {
			s.mu.Lock()
			s.flushDelta(id, task)
			task.ToolCalls = append(task.ToolCalls, shared.ToolCall{
				Name:       evt.ToolName,
				ToolID:     evt.ToolID,
				Status:     "pending",
				Parameters: evt.Parameters,
				Timestamp:  time.Now(),
			})
			s.mu.Unlock()
		}
	})

	exec.On("tool_result", func(data interface{}) {
		if evt, ok := data.(shared.CliToolResultEvent); ok {
			s.mu.Lock()
			defer s.mu.Unlock()
			for i := range task.ToolCalls {
				if task.ToolCalls[i].ToolID == evt.ToolID {
					task.ToolCalls[i].Status = evt.Status
					task.ToolCalls[i].Output = evt.Output
					break
				}
			}
		}
	})

	exec.On("result", func(data interface{}) {
		if evt, ok := data.(shared.CliResultEvent); ok {
			s.mu.Lock()
			s.flushDelta(id, task)
			if evt.Status == "success" {
				task.State = shared.TaskStateCompleted
			} else {
				task.State = shared.TaskStateFailed
			}
			completedAt := time.Now()
			task.CompletedAt = &completedAt
			task.Result = &shared.TaskResult{
				FinalResponse: s.lastAssistantMsg(task),
				Stats:         evt.Stats,
			}
			s.mu.Unlock()
		}
	})

	exec.On("error", func(data interface{}) {
		if err, ok := data.(error); ok {
			s.mu.Lock()
			task.State = shared.TaskStateFailed
			task.Error = err.Error()
			completedAt := time.Now()
			task.CompletedAt = &completedAt
			s.mu.Unlock()
		}
	})

	exec.On("timeout", func(data interface{}) {
		s.mu.Lock()
		task.State = shared.TaskStateTimeout
		task.Error = "Task timed out"
		completedAt := time.Now()
		task.CompletedAt = &completedAt
		s.mu.Unlock()
	})

	exec.On("exit", func(data interface{}) {
		s.mu.Lock()
		delete(s.executors, id)
		delete(s.deltaBuffers, id)
		if task.State == shared.TaskStateRunning {
			task.State = shared.TaskStateFailed
			task.Error = "Process exited unexpectedly"
			completedAt := time.Now()
			task.CompletedAt = &completedAt
		}
		s.mu.Unlock()
	})
}

func (s *Scheduler) flushDelta(id string, task *shared.TaskRecord) {
	buf := s.deltaBuffers[id]
	if buf != "" {
		task.Messages = append(task.Messages, shared.Message{
			Role:      "assistant",
			Content:   buf,
			Timestamp: time.Now(),
		})
		s.deltaBuffers[id] = ""
	}
}

func (s *Scheduler) lastAssistantMsg(task *shared.TaskRecord) string {
	for i := len(task.Messages) - 1; i >= 0; i-- {
		if task.Messages[i].Role == "assistant" {
			return task.Messages[i].Content
		}
	}
	return ""
}

func (s *Scheduler) sweep() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, task := range s.tasks {
		if task.State != shared.TaskStateRunning && task.CompletedAt != nil {
			if now.Sub(*task.CompletedAt) > shared.STALE_TASK_AGE_MS*time.Millisecond {
				delete(s.tasks, id)
				delete(s.executors, id)
				delete(s.deltaBuffers, id)
			}
		}
	}

	s.sweepTimer = time.AfterFunc(shared.CLEANUP_INTERVAL_MS*time.Millisecond, func() {
		s.sweep()
	})
}

func generateID() string {
	return uuid.New().String()[:8]
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
