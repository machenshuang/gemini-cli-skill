package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "cli-agent",
	Short: "CLI task manager for Gemini and Kimi",
	Long: `cli-agent — CLI task manager for Gemini and Kimi

A daemon-based task runner that manages AI assistant tasks.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", true, "Output in JSON format (default)")
}

// die 打印错误并退出
func die(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}

// jsonOutputData 输出 JSON 数据
func jsonOutputData(data interface{}) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		die(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	fmt.Println(string(bytes))
}
