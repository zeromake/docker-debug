package command

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zeromake/docker-debug/internal/config"
	"github.com/zeromake/docker-debug/pkg/tty"
)

var rootCmd = newExecCommand()

type execOptions struct {
	host       string
	image      string
	detachKeys string
	user       string
	privileged bool
	workDir    string
	targetDir  string
	container  string
	certDir    string
	command    []string
	name string
}

func newExecOptions() execOptions {
	return execOptions{}
}

func newExecCommand() *cobra.Command {
	options := newExecOptions()

	cmd := &cobra.Command{
		Use:   "docker-debug [OPTIONS] CONTAINER COMMAND [ARG...]",
		Short: "Run a command in a running container",
		Args:  RequiresMinArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.container = args[0]
			options.command = args[1:]
			return runExec(options)
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	flags.StringVarP(&options.image, "image", "i", "", "use this image")
	flags.StringVarP(&options.name, "name", "n", "", "docker config name")
	flags.StringVarP(&options.host, "host", "H", "", "connection host's docker (format: tcp://192.168.99.100:2376)")
	flags.StringVarP(&options.certDir, "cert-dir", "c", "", "cert dir use tls")
	flags.StringVarP(&options.detachKeys, "detach-keys", "d", "", "Override the key sequence for detaching a container")
	flags.StringVarP(&options.user, "user", "u", "", "Username or UID (format: <name|uid>[:<group|gid>])")
	flags.BoolVarP(&options.privileged, "privileged", "p", false, "Give extended privileges to the command")
	flags.StringVarP(&options.workDir, "work-dir", "w", "", "Working directory inside the container")
	_ = flags.SetAnnotation("work-dir", "version", []string{"1.35"})
	flags.StringVarP(&options.targetDir, "target-dir", "t", "", "Working directory inside the container")
	return cmd
}

func runExec(options execOptions) error {
	logrus.SetLevel(logrus.ErrorLevel)
	var containerId string
	conf, err := config.LoadConfig()
	opts := []DebugCliOption{
		WithConfig(conf),
	}
	if options.image != "" {
		conf.Image = options.image
	}
	if conf.Image == "" {
		return errors.New("not set image!")
	}
	if options.host != "" {
		dockerConfig := config.DockerConfig{
			Host: options.host,
		}
		if options.certDir != "" {
			dockerConfig.TLS = true
			dockerConfig.CertDir = options.certDir
		}
		opts = append(opts, WithClientConfig(dockerConfig))
	} else {
		name := conf.DockerConfigDefault
		opt, ok := conf.DockerConfig[conf.DockerConfigDefault]
		if options.name != "" {
			name = options.name
			opt, ok = conf.DockerConfig[options.name]
		}
		if !ok {
			return errors.Errorf("not find %s docker config", name)
		}
		opts = append(opts, WithClientConfig(opt))
	}

	cli, err := NewDebugCli(opts...)
	if err != nil {
		return err
	}
	defer func() {
		if containerId != "" {
			err = cli.ContainerClean(containerId)
			if err != nil {
				logrus.Debugf("%+v", err)
			}
		}
		err = cli.Close()
		if err != nil {
			logrus.Debugf("%+v", err)
		}
	}()
	_, err = cli.Ping()
	if err != nil {
		return err
	}
	// find image
	images, err := cli.FindImage(conf.Image)
	if err != nil {
		return err
	}

	if len(images) == 0 {
		// pull image
		err = cli.PullImage(conf.Image)
		if err != nil {
			return err
		}
	}
	containerId, err = cli.FindContainer(options.container)
	if err != nil {
		return err
	}
	containerId, err = cli.CreateContainer(containerId)
	if err != nil {
		return err
	}
	resp, err := cli.ExecCreate(options, containerId)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		errCh <- func() error {
			return cli.ExecStart(options, resp.ID)
		}()
	}()
	if cli.In().IsTerminal() {
		if err := tty.MonitorTtySize(context.Background(), cli.Client(), cli.Out(), resp.ID, true); err != nil {
			_, _ = fmt.Fprintln(cli.Err(), "Error monitoring TTY size:", err)
		}
	}

	if err := <-errCh; err != nil {
		logrus.Debugf("Error hijack: %s", err)
		return err
	}
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Debugf("%+v", err)
	}
}
