package cmd

import (
	"github.com/spf13/cobra"
)

var accountName string

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates resources",
	Long:  `Creates resources in EPIC.`,
}

func init() {
	createCmd.PersistentFlags().StringVar(&accountName, "account-name", "root", "name of the user account")
	rootCmd.AddCommand(createCmd)
}
