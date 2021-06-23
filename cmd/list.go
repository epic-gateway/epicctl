package cmd

import (
	"github.com/spf13/cobra"
)

// listCmd is a container command for the subcommands that list
// various types of resources.
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "Lists resources",
	Long:    `Lists resources in EPIC.`,
}

func init() {
	rootCmd.AddCommand(listCmd)
}
