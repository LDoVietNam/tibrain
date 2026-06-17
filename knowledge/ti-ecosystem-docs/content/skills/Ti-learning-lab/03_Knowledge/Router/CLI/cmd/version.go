package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	version = "2.0.0-go123"
	commit  = "dev"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print build version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ti-cli %s\ncommit: %s\nbuilt: %s\ngo: %s\nplatform: %s/%s\n", version, commit, date, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		if info, ok := debug.ReadBuildInfo(); ok {
			fmt.Printf("module: %s\n", info.Main.Path)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
