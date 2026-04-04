package main

import (
	"cli-agent-go/cmd"
	"cli-agent-go/daemon"
	"flag"
	"fmt"
	"os"
)

func main() {
	// Check for hidden --daemon-mode flag
	var daemonMode bool
	flag.BoolVar(&daemonMode, "daemon-mode", false, "Run in daemon mode (internal)")
	flag.Parse()

	if daemonMode {
		// Run as daemon server
		if err := daemon.StartServer(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start daemon: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Run as CLI
	cmd.Execute()
}
