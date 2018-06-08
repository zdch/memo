package cmd

import (
	"github.com/memocash/memo/app/bitcoin/main-node"
	"github.com/spf13/cobra"
	"os"
	"fmt"
	"time"
)

var mainNodeCmd = &cobra.Command{
	Use: "main-node",
	RunE: func(c *cobra.Command, args []string) error {
		var last time.Time
		for last.IsZero() || time.Since(last) > time.Minute {
			last = time.Now()
			main_node.Start()
			main_node.WaitForDisconnect()
			fmt.Println("Disconnected.")
		}
		os.Exit(1)
		return nil
	},
}
