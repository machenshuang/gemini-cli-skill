package cmd

import (
	"cli-agent-go/daemon"
	"cli-agent-go/shared"
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd list 命令
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all tasks",
	Long:    `List all tasks with optional filtering by state and tags.`,
	Example: `  cli-agent list
  cli-agent list --state running --state pending
  cli-agent list --tag review --limit 10`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if daemon is running
		if err := daemon.CheckDaemonRunning(); err != nil {
			die(err.Error())
		}

		states, _ := cmd.Flags().GetStringArray("state")
		tags, _ := cmd.Flags().GetStringArray("tag")
		limit, _ := cmd.Flags().GetInt("limit")

		params := map[string]interface{}{}

		if len(states) > 0 {
			var stateList []string
			for _, s := range states {
				stateList = append(stateList, s)
			}
			params["state"] = stateList
		}

		if len(tags) > 0 {
			params["tags"] = tags
		}

		if limit > 0 {
			params["limit"] = limit
		}

		res, err := daemon.Rpc(shared.RpcRequest{
			Action: shared.RpcActionList,
			Params: params,
		})
		if err != nil {
			die(fmt.Sprintf("failed to list tasks: %v", err))
		}

		if !res.Ok {
			die(res.Error)
		}

		jsonOutputData(res.Data)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringArray("state", []string{}, "Filter by state (repeatable)")
	listCmd.Flags().StringArray("tag", []string{}, "Filter by tag (repeatable)")
	listCmd.Flags().Int("limit", 20, "Max results (default: 20)")
}
