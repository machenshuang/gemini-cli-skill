package daemon

import (
	"bufio"
	"cli-agent-go/shared"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

// IsDaemonRunning 检查守护进程是否运行
func IsDaemonRunning() bool {
	if _, err := os.Stat(shared.SOCKET_PATH); os.IsNotExist(err) {
		return false
	}

	conn, err := net.DialTimeout("unix", shared.SOCKET_PATH, 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Rpc 发送 RPC 请求到守护进程
func Rpc(req shared.RpcRequest) (*shared.RpcResponse, error) {
	conn, err := net.DialTimeout("unix", shared.SOCKET_PATH, shared.CONNECT_TIMEOUT_MS*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to daemon: %w", err)
	}
	defer conn.Close()

	// Set write deadline
	conn.SetWriteDeadline(time.Now().Add(shared.CONNECT_TIMEOUT_MS * time.Millisecond))

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if _, err := conn.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(shared.RESPONSE_TIMEOUT_MS * time.Millisecond))

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("daemon response timeout or error: %w", err)
	}

	line = line[:len(line)-1] // Remove newline

	var res shared.RpcResponse
	if err := json.Unmarshal([]byte(line), &res); err != nil {
		return nil, fmt.Errorf("invalid response from daemon: %w", err)
	}

	return &res, nil
}

// CheckDaemonRunning 检查守护进程是否在运行，如果没有则返回错误
func CheckDaemonRunning() error {
	if !IsDaemonRunning() {
		return fmt.Errorf("daemon is not running. Please start it with: cli-agent daemon start")
	}
	return nil
}
