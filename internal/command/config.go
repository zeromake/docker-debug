package command

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zeromake/docker-debug/internal/config"
)

func init() {
	cfg := &config.DockerConfig{}
	name := ""
	cmd := &cobra.Command{
		Use:   "config",
		Short: "docker conn config cli",
		Args:  RequiresMinArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.LoadConfig()
			if err != nil {
				return err
			}
			if cfg.Host == "" {
				c, ok := conf.DockerConfig[name]
				if ok {
					fmt.Printf("config `%s`:\n%+v\n", name, c)
					return nil
				}
				return errors.Errorf("not find %s config", name)
			}
			conf.DockerConfig[name] = cfg
			return conf.Save()
		},
	}
	flags := cmd.Flags()
	flags.SetInterspersed(false)
	flags.StringVarP(&name, "name", "n", "default", "docker config name")
	flags.BoolVarP(&cfg.TLS, "tls", "t", false, "docker conn is tls")
	flags.StringVarP(&cfg.CertDir, "cert-dir", "c", "", "docker tls cert dir")
	flags.StringVarP(&cfg.Host, "host", "H", "", "docker host")
	flags.StringVarP(&cfg.CertPassword, "password", "p", "", "docker tls password")
	rootCmd.AddCommand(cmd)
}
