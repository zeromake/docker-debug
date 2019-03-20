package command

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zeromake/docker-debug/version"
)

func init()  {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "docker and client info",
		Args:  RequiresMinArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Version:\t%s\n", version.Version)
			fmt.Printf("Platform:\t%s\n", version.PlatformName)
			fmt.Printf("Commit:\t\t%s\n", version.GitCommit)
			fmt.Printf("Time:\t\t%s\n", version.BuildTime)
			return nil
		},
	}
	rootCmd.AddCommand(cmd)
}
