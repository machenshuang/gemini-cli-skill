package cmd

import (
	"cli-agent-go/daemon"
	"cli-agent-go/shared"
	"fmt"

	"github.com/spf13/cobra"
)

// statusCmd status 命令
var statusCmd = &cobra.Command{
	Use:   "status [task_id]",
	Short: "Query task status and output",
	Long:  `Query the status and output of a running or completed task.`,
	Example: `  cli-agent status abc123
  cli-agent status abc123 --verbosity full
  cli-agent status abc123 --tail 10`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if daemon is running
		if err := daemon.CheckDaemonRunning(); err != nil {
			die(err.Error())
		}

		taskID := args[0]
		verbosity, _ := cmd.Flags().GetString("verbosity")
		tail, _ := cmd.Flags().GetInt("tail")

		if verbosity == "" {
			verbosity = "normal"
		}

		res, err := daemon.Rpc(shared.RpcRequest{
			Action: shared.RpcActionStatus,
			Params: map[string]interface{}{
				"taskId":    taskID,
				"verbosity": verbosity,
				"tail":      tail,
			},
		})
		if err != nil {
			die(fmt.Sprintf("failed to get status: %v", err))
		}

		if !res.Ok {
			die(res.Error)
		}

		jsonOutputData(res.Data)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringP("verbosity", "v", "normal", "Output verbosity: minimal | normal | full")
	statusCmd.Flags().Int("tail", 0, "Show last N messages only")
}
