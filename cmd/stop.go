package cmd

import (
	"cli-agent-go/daemon"
	"cli-agent-go/shared"
	"fmt"

	"github.com/spf13/cobra"
)

// stopCmd stop 命令
var stopCmd = &cobra.Command{
	Use:   "stop [task_id]",
	Short: "Stop a running task",
	Long:  `Stop a running task gracefully or forcefully.`,
	Example: `  cli-agent stop abc123
  cli-agent stop abc123 --force`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if daemon is running
		if err := daemon.CheckDaemonRunning(); err != nil {
			die(err.Error())
		}

		taskID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		res, err := daemon.Rpc(shared.RpcRequest{
			Action: shared.RpcActionStop,
			Params: map[string]interface{}{
				"taskId": taskID,
				"force":  force,
			},
		})
		if err != nil {
			die(fmt.Sprintf("failed to stop task: %v", err))
		}

		if !res.Ok {
			die(res.Error)
		}

		jsonOutputData(res.Data)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	stopCmd.Flags().BoolP("force", "f", false, "Force kill (SIGKILL)")
}
