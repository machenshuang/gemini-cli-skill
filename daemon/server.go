package daemon

import (
	"bufio"
	"cli-agent-go/engine"
	"cli-agent-go/shared"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Server 守护进程服务器
type Server struct {
	listener   net.Listener
	scheduler  *engine.Scheduler
	idleTimer  *time.Timer
	mu         sync.Mutex
}

var serverInstance *Server

// StartServer 启动守护进程服务器
func StartServer() error {
	cfg := shared.LoadConfig()
	
	serverInstance = &Server{
		scheduler: engine.NewScheduler(cfg),
	}

	// Ensure base dir exists
	if err := os.MkdirAll(shared.BASE_DIR, 0755); err != nil {
		return fmt.Errorf("failed to create base dir: %w", err)
	}

	// Clean stale socket
	os.Remove(shared.SOCKET_PATH)

	listener, err := net.Listen("unix", shared.SOCKET_PATH)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}
	serverInstance.listener = listener

	// Write PID file
	pid := os.Getpid()
	if err := os.WriteFile(shared.PID_PATH, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	fmt.Printf("cli-agent daemon started (pid=%d)\n", pid)

	// Reset idle timer
	serverInstance.resetIdleTimer()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		serverInstance.Stop()
	}()

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			if serverInstance.listener == nil {
				return nil // Server stopped
			}
			continue
		}
		go serverInstance.handleConnection(conn)
	}
}

// Stop 停止守护进程
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.idleTimer != nil {
		s.idleTimer.Stop()
		s.idleTimer = nil
	}

	s.scheduler.Shutdown()

	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}

	os.Remove(shared.SOCKET_PATH)
	os.Remove(shared.PID_PATH)

	fmt.Println("Daemon stopped.")
	os.Exit(0)
}

func (s *Server) resetIdleTimer() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.idleTimer != nil {
		s.idleTimer.Stop()
	}

	s.idleTimer = time.AfterFunc(shared.DAEMON_IDLE_TIMEOUT_MS*time.Millisecond, func() {
		if s.scheduler.RunningCount() > 0 {
			// Still has running tasks, check again later
			s.resetIdleTimer()
			return
		}
		fmt.Println("Daemon idle for 30 minutes with no running tasks. Shutting down.")
		s.Stop()
	})
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = line[:len(line)-1] // Remove newline
		if line == "" {
			continue
		}

		s.resetIdleTimer()
		s.processRequest(conn, line)
	}
}

func (s *Server) processRequest(conn net.Conn, raw string) {
	var req shared.RpcRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		s.reply(conn, shared.RpcResponse{Ok: false, Error: "Invalid JSON"})
		return
	}

	data, err := s.route(req)
	if err != nil {
		s.reply(conn, shared.RpcResponse{Ok: false, Error: err.Error()})
		return
	}

	s.reply(conn, shared.RpcResponse{Ok: true, Data: data})
}

func (s *Server) route(req shared.RpcRequest) (interface{}, error) {
	switch req.Action {
	case shared.RpcActionStart:
		return s.handleStart(req.Params)
	case shared.RpcActionStatus:
		return s.handleStatus(req.Params)
	case shared.RpcActionStop:
		return s.handleStop(req.Params)
	case shared.RpcActionList:
		return s.handleList(req.Params)
	case shared.RpcActionShutdown:
		// Defer stop so reply goes out first
		go func() {
			time.Sleep(100 * time.Millisecond)
			s.Stop()
		}()
		return map[string]string{"message": "Shutting down"}, nil
	default:
		return nil, fmt.Errorf("unknown action: %s", req.Action)
	}
}

func (s *Server) reply(conn net.Conn, res shared.RpcResponse) {
	data, err := json.Marshal(res)
	if err != nil {
		return
	}
	conn.Write(append(data, '\n'))
}

func (s *Server) handleStart(params map[string]interface{}) (interface{}, error) {
	startParams := shared.StartParams{
		Prompt:     getStringParam(params, "prompt"),
		WorkingDir: getStringParam(params, "workingDir"),
		Model:      getStringParam(params, "model"),
		Timeout:    getIntParam(params, "timeout"),
		Tags:       getStringSliceParam(params, "tags"),
	}

	if v, ok := params["approvalMode"].(string); ok {
		startParams.ApprovalMode = shared.ApprovalMode(v)
	}
	if v, ok := params["backend"].(string); ok {
		startParams.Backend = shared.Backend(v)
	}
	if v, ok := params["thinking"].(bool); ok {
		startParams.Thinking = v
	}

	task, err := s.scheduler.CreateTask(startParams)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"task_id":    task.ID,
		"session_id": task.SessionID,
		"state":      task.State,
		"started_at": task.StartedAt,
		"backend":    task.Backend,
	}, nil
}

func (s *Server) handleStatus(params map[string]interface{}) (interface{}, error) {
	taskID := getStringParam(params, "taskId")
	if taskID == "" {
		return nil, fmt.Errorf("taskId is required")
	}

	task := s.scheduler.GetTask(taskID)
	if task == nil {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	verbosity := shared.Verbosity(getStringParam(params, "verbosity"))
	if verbosity == "" {
		verbosity = shared.VerbosityNormal
	}
	tail := getIntParam(params, "tail")

	maxLen := 500
	if verbosity == shared.VerbosityFull {
		maxLen = 0
	}

	cut := func(s string) string {
		if maxLen == 0 || len(s) <= maxLen {
			return s
		}
		return s[:maxLen] + "... [truncated]"
	}

	msgs := task.Messages
	tcs := task.ToolCalls
	if tail > 0 {
		if len(msgs) > tail {
			msgs = msgs[len(msgs)-tail:]
		}
		if len(tcs) > tail {
			tcs = tcs[len(tcs)-tail:]
		}
	}

	elapsed := time.Since(task.StartedAt).Milliseconds()
	if task.CompletedAt != nil {
		elapsed = task.CompletedAt.Sub(task.StartedAt).Milliseconds()
	}

	res := map[string]interface{}{
		"task_id": task.ID,
		"state":   task.State,
		"backend": task.Backend,
		"progress": map[string]interface{}{
			"messages":   len(task.Messages),
			"tool_calls": len(task.ToolCalls),
			"elapsed_ms": elapsed,
		},
	}

	if verbosity != shared.VerbosityMinimal {
		var msgList []map[string]interface{}
		for _, m := range msgs {
			msgList = append(msgList, map[string]interface{}{
				"role":      m.Role,
				"content":   cut(m.Content),
				"timestamp": m.Timestamp,
			})
		}

		var tcList []map[string]interface{}
		for _, tc := range tcs {
			tcData := map[string]interface{}{
				"name":       tc.Name,
				"tool_id":    tc.ToolID,
				"status":     tc.Status,
				"output":     cut(tc.Output),
				"timestamp":  tc.Timestamp,
			}
			if verbosity == shared.VerbosityFull {
				tcData["parameters"] = tc.Parameters
			}
			tcList = append(tcList, tcData)
		}

		res["output"] = map[string]interface{}{
			"messages":   msgList,
			"tool_calls": tcList,
		}
	}

	if task.Result != nil {
		res["result"] = map[string]interface{}{
			"final_response": cut(task.Result.FinalResponse),
			"stats":          task.Result.Stats,
		}
	}
	if task.Error != "" {
		res["error"] = task.Error
	}

	return res, nil
}

func (s *Server) handleStop(params map[string]interface{}) (interface{}, error) {
	taskID := getStringParam(params, "taskId")
	if taskID == "" {
		return nil, fmt.Errorf("taskId is required")
	}

	force := getBoolParam(params, "force")

	ok := s.scheduler.StopTask(taskID, force)
	if !ok {
		return nil, fmt.Errorf("cannot stop task %s (not running or not found)", taskID)
	}

	task := s.scheduler.GetTask(taskID)
	return map[string]interface{}{
		"success": true,
		"state":   task.State,
	}, nil
}

func (s *Server) handleList(params map[string]interface{}) (interface{}, error) {
	filter := &shared.ListParams{}

	if v, ok := params["limit"].(float64); ok {
		filter.Limit = int(v)
	}

	if v, ok := params["tags"].([]interface{}); ok {
		for _, t := range v {
			if s, ok := t.(string); ok {
				filter.Tags = append(filter.Tags, s)
			}
		}
	}

	if v, ok := params["state"].([]interface{}); ok {
		for _, s := range v {
			if str, ok := s.(string); ok {
				filter.State = append(filter.State, shared.TaskState(str))
			}
		}
	}

	tasks := s.scheduler.ListTasks(filter)
	return map[string]interface{}{
		"tasks":   tasks,
		"total":   s.scheduler.TotalCount(),
		"running": s.scheduler.RunningCount(),
	}, nil
}

func getStringParam(params map[string]interface{}, key string) string {
	if v, ok := params[key].(string); ok {
		return v
	}
	return ""
}

func getIntParam(params map[string]interface{}, key string) int {
	if v, ok := params[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getBoolParam(params map[string]interface{}, key string) bool {
	if v, ok := params[key].(bool); ok {
		return v
	}
	return false
}

func getStringSliceParam(params map[string]interface{}, key string) []string {
	var result []string
	if v, ok := params[key].([]interface{}); ok {
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
	}
	return result
}
