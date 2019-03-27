package command

import (
	"github.com/spf13/cobra"
	"github.com/zeromake/docker-debug/internal/config"
)

func init() {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "docker-debug init config",
		Args:  RequiresMinArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := config.InitConfig()
			return err
		},
	}
	rootCmd.AddCommand(cmd)
}
