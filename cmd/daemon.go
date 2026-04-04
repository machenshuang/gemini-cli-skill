package cmd

import (
	"cli-agent-go/daemon"
	"cli-agent-go/shared"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// daemonCmd daemon 命令
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the daemon",
	Long:  `Start, stop, or check the status of the cli-agent daemon.`,
}

// daemonStartCmd daemon start 子命令
var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon",
	Long:  `Start the cli-agent daemon in the background.`,
	Run: func(cmd *cobra.Command, args []string) {
		foreground, _ := cmd.Flags().GetBool("foreground")

		if foreground {
			// Run in foreground
			if err := daemon.StartServer(); err != nil {
				die(fmt.Sprintf("failed to start daemon: %v", err))
			}
			return
		}

		// Check if already running
		if daemon.IsDaemonRunning() {
			fmt.Println("Daemon is already running.")
			return
		}

		// Start daemon in background
		exe, err := os.Executable()
		if err != nil {
			// Try to find the binary in PATH
			exe = "cli-agent"
		}

		// Get the absolute path
		exe, _ = filepath.Abs(exe)

		procAttr := &os.ProcAttr{
			Dir:   ".",
			Env:   os.Environ(),
			Files: []*os.File{nil, nil, nil}, // stdin, stdout, stderr
			Sys: &syscall.SysProcAttr{
				Setsid: true,
			},
		}

		process, err := os.StartProcess(exe, []string{filepath.Base(exe), "--daemon-mode"}, procAttr)
		if err != nil {
			// Fallback: use exec.Command with Start
			cmd_ := exec.Command(exe, "--daemon-mode")
			cmd_.SysProcAttr = &syscall.SysProcAttr{
				Setsid: true,
			}
			cmd_.Stdin = nil
			cmd_.Stdout = nil
			cmd_.Stderr = nil
			if err := cmd_.Start(); err != nil {
				die(fmt.Sprintf("failed to start daemon: %v", err))
			}
			fmt.Printf("Daemon starting in background (pid=%d)\n", cmd_.Process.Pid)
			return
		}

		fmt.Printf("Daemon starting in background (pid=%d)\n", process.Pid)
		process.Release()
	},
}

// daemonStopCmd daemon stop 子命令
var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon",
	Long:  `Stop the running cli-agent daemon.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !daemon.IsDaemonRunning() {
			fmt.Println("Daemon is not running.")
			return
		}

		res, err := daemon.Rpc(shared.RpcRequest{
			Action: shared.RpcActionShutdown,
			Params: map[string]interface{}{},
		})
		if err != nil {
			die(fmt.Sprintf("failed to stop daemon: %v", err))
		}

		if !res.Ok {
			die(res.Error)
		}

		fmt.Println("Daemon shutdown requested.")
	},
}

// daemonStatusCmd daemon status 子命令
var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check daemon status",
	Long:  `Check if the cli-agent daemon is running.`,
	Run: func(cmd *cobra.Command, args []string) {
		if daemon.IsDaemonRunning() {
			fmt.Println("Daemon is running.")
		} else {
			fmt.Println("Daemon is not running.")
		}
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)

	daemonStartCmd.Flags().BoolP("foreground", "f", false, "Run daemon in foreground")
}

// WaitForDaemon 等待守护进程启动
func WaitForDaemon(timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if daemon.IsDaemonRunning() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}
