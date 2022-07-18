package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// version is set by the build process.
var version string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Epicctl version",
	Long:    `Outputs this program's version info.`,
	Run: func(cmd *cobra.Command, args []string) {
		if info, ok := debug.ReadBuildInfo(); ok {
			Debug("%s\n", info)
		}
	},
}

func debugVersion() {
	fmt.Fprintf(os.Stderr, "%s %s\n", "epicctl", version)
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
