package command

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zeromake/docker-debug/internal/config"
)

func init() {
	cmd := &cobra.Command{
		Use:   "use",
		Short: "docker set default config",
		Args:  RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.LoadConfig()
			if err != nil {
				return err
			}
			name := args[0]
			_, ok := conf.DockerConfig[name]
			if !ok {
				return errors.Errorf("not find %s config", name)
			}
			conf.DockerConfigDefault = name
			return conf.Save()
		},
	}
	rootCmd.AddCommand(cmd)
}
