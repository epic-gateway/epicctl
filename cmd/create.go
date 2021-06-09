package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resource",
	Long:  `Create a resource in EPIC.`,
}

var serviceprefixCmd = &cobra.Command{
	Use:   "serviceprefix",
	Short: "Create Service Prefix",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("serviceprefix called")
	},
}
var lbgCmd = &cobra.Command{
	Use:   "lbg",
	Short: "Create Load Balancer Group",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("lbg called")
	},
}
var envoytemplateCmd = &cobra.Command{
	Use:   "envoytemplate",
	Short: "Create Envoy Template",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("envoytemplate called")
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(serviceprefixCmd)
	createCmd.AddCommand(lbgCmd)
	createCmd.AddCommand(envoytemplateCmd)
}
