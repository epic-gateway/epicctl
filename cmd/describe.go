package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(describeCmd)
}

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describes resources",
	Long: `Describes resources in the EPIC system.

This command's subcommands describe different types of resources.`,
}
