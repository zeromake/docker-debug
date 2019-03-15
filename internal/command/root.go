package command

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zeromake/docker-debug/internal/config"
	"os"
)

var rootCmd = newExecCommand()

type execOptions struct {
	host 		string
	image       string
	detachKeys  string
	user        string
	privileged  bool
	workdir     string
	container   string
	command     []string
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

	flags.StringVarP(&options.image, "image", "", "", "use this image")
	flags.StringVarP(&options.host, "host", "", "", "conn this host's docker (format: tcp://192.168.99.100:2376)")
	flags.StringVarP(&options.detachKeys, "detach-keys", "", "", "Override the key sequence for detaching a container")
	flags.StringVarP(&options.user, "user", "u", "", "Username or UID (format: <name|uid>[:<group|gid>])")
	flags.BoolVarP(&options.privileged, "privileged", "", false, "Give extended privileges to the command")
	flags.StringVarP(&options.workdir, "workdir", "w", "", "Working directory inside the container")
	_ = flags.SetAnnotation("workdir", "version", []string{"1.35"})
	return cmd
}

func runExec(options execOptions) error {
	conf, err := config.LoadConfig()
	opts := []DebugCliOption{
		WithConfig(conf),
	}
	if options.image != "" {
		conf.Image = options.image
	}
	if options.host != "" {
		dockerConfig := config.DockerConfig{
			Host: options.host,
		}
		opts = append(opts, WithClientConfig(dockerConfig))
	} else {
		opts = append(opts, WithClientName(conf.DockerConfigDefault))
	}

	cli, err := NewDebugCli(opts...)
	if err != nil {
		return err
	}
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
	containerId, err := cli.FindContainer(options.container)
	if err != nil {
		return err
	}
	containerId, err = cli.CreateContainer(containerId)
	if err != nil {
		return err
	}
	defer func() {
		_ = cli.ContainerClean(containerId)
	}()
	resp, err := cli.ExecCreate(options, containerId)
	if err != nil {
		return err
	}
	return cli.ExecStart(options, resp.ID)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}

