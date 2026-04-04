package cmd

import (
	"cli-agent-go/daemon"
	"cli-agent-go/shared"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// startCmd start 命令
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new task",
	Long: `Start a new AI assistant task.

The daemon must be running. Use 'cli-agent daemon start' to start it.`,
	Example: `  cli-agent start -p "Write a hello world program in Go"
  cli-agent start -p "Review this code" -C /path/to/project --tag review
  cli-agent start -p "Explain this function" -b kimi -m gpt-4`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if daemon is running
		if err := daemon.CheckDaemonRunning(); err != nil {
			die(err.Error())
		}

		prompt, _ := cmd.Flags().GetString("prompt")
		if prompt == "" {
			die("Missing required flag: -p, --prompt")
		}

		workingDir, _ := cmd.Flags().GetString("cwd")
		if workingDir == "" {
			workingDir, _ = os.Getwd()
		}

		model, _ := cmd.Flags().GetString("model")
		approvalMode, _ := cmd.Flags().GetString("approval-mode")
		timeout, _ := cmd.Flags().GetInt("timeout")
		tags, _ := cmd.Flags().GetStringArray("tag")
		backend, _ := cmd.Flags().GetString("backend")
		thinking, _ := cmd.Flags().GetBool("thinking")
		noThinking, _ := cmd.Flags().GetBool("no-thinking")
		_ = noThinking // Unused but kept for flag parsing

		cfg := shared.LoadConfig()

		// Set defaults from config
		if backend == "" {
			backend = string(cfg.DefaultBackend)
		}
		if approvalMode == "" {
			approvalMode = string(cfg.DefaultApprovalMode)
		}
		if timeout == 0 {
			timeout = cfg.DefaultTimeout
		}

		// Parse thinking flag
		var thinkingVal bool
		if cmd.Flags().Changed("thinking") {
			thinkingVal = thinking
		} else if cmd.Flags().Changed("no-thinking") {
			thinkingVal = false
		} else {
			thinkingVal = cfg.DefaultThinking
		}

		params := shared.StartParams{
			Prompt:       prompt,
			WorkingDir:   workingDir,
			Model:        model,
			ApprovalMode: shared.ApprovalMode(approvalMode),
			Timeout:      timeout,
			Tags:         tags,
			Backend:      shared.Backend(backend),
			Thinking:     thinkingVal,
		}

		res, err := daemon.Rpc(shared.RpcRequest{
			Action: shared.RpcActionStart,
			Params: map[string]interface{}{
				"prompt":       params.Prompt,
				"workingDir":   params.WorkingDir,
				"model":        params.Model,
				"approvalMode": string(params.ApprovalMode),
				"timeout":      params.Timeout,
				"tags":         params.Tags,
				"backend":      string(params.Backend),
				"thinking":     params.Thinking,
			},
		})
		if err != nil {
			die(fmt.Sprintf("failed to start task: %v", err))
		}

		if !res.Ok {
			die(res.Error)
		}

		jsonOutputData(res.Data)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringP("prompt", "p", "", "Task prompt (required)")
	startCmd.Flags().StringP("cwd", "C", "", "Working directory")
	startCmd.Flags().StringP("model", "m", "", "Model name (backend-specific)")
	startCmd.Flags().StringP("approval-mode", "a", "", "Approval mode: default | auto_edit | yolo")
	startCmd.Flags().Int("timeout", 0, "Timeout in seconds (0 = no timeout)")
	startCmd.Flags().StringArray("tag", []string{}, "Add tag (repeatable)")
	startCmd.Flags().StringP("backend", "b", "", "Backend: gemini | kimi (default: from config or kimi)")
	startCmd.Flags().Bool("thinking", false, "Enable thinking mode (Kimi only)")
	startCmd.Flags().Bool("no-thinking", false, "Disable thinking mode (Kimi only)")
}
