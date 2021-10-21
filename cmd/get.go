package cmd

import (
	"github.com/spf13/cobra"
)

// getCmd is a container command for the subcommands that list
// various types of resources.
var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Gets resources",
	Long:    `Gets resources in EPIC.`,
}

func init() {
	rootCmd.AddCommand(getCmd)
}
