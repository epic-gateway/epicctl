package cmd

import (
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"del", "d"},
	Short:   "Deletes resources",
	Long:    `Deletes resources from EPIC.`,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
