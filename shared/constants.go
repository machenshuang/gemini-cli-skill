package shared

import (
	"os"
	"path/filepath"
)

var (
	// Base directory for all cli-agent data
	HomeDir, _ = os.UserHomeDir()
	BASE_DIR   = filepath.Join(HomeDir, ".cli-agent")

	// Daemon socket and PID
	SOCKET_PATH = filepath.Join(BASE_DIR, "daemon.sock")
	PID_PATH    = filepath.Join(BASE_DIR, "daemon.pid")

	// Task state files (standalone mode - not used in Go version)
	TASKS_DIR = filepath.Join(BASE_DIR, "tasks")

	// Configuration
	CONFIG_PATH = filepath.Join(BASE_DIR, "config.json")
)

// Defaults
const (
	DEFAULT_MAX_CONCURRENT   = 3
	DEFAULT_TIMEOUT          = 600 // seconds
	DEFAULT_APPROVAL_MODE    = ApprovalModeAutoEdit
	DEFAULT_BACKEND          = BackendKimi
	TASK_ID_LENGTH           = 8
	STALE_TASK_AGE_MS        = 24 * 60 * 60 * 1000 // 24 hours
	CLEANUP_INTERVAL_MS      = 30_000
	DAEMON_IDLE_TIMEOUT_MS   = 30 * 60 * 1000 // 30 minutes
	CONNECT_TIMEOUT_MS       = 3000
	RESPONSE_TIMEOUT_MS      = 30000
)
