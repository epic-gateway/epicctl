package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func init() {
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "markdown",
			Short: "Create markdown docs",
			Long: `Create markdown docs for epicctl's commands.

One parameter is required: the path to the directory in which the docs
will be generated.
`,
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return doc.GenMarkdownTree(rootCmd, args[0])
			},
		},
	)
}
