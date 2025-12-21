package cmd

import (
	"fmt"
	"os"
	
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "subconverter",
	Short: "A subscription converter and merger tool",
	Long: `Subconverter is a tool to merge multiple proxy subscriptions 
and convert them to different formats (Base64, Clash, etc.)`,
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
