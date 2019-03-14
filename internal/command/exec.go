package command

import (
	"github.com/spf13/cobra"
)

type execOptions struct {
	image       string
	detachKeys  string
	interactive bool
	tty         bool
	user        string
	privileged  bool
	workdir     string
	container   string
	command     []string
}

func newExecOptions() execOptions {
	return execOptions{}
}

func NewExecCommand(dockerCli Cli) *cobra.Command {
	options := newExecOptions()

	cmd := &cobra.Command{
		Use:   "docker-debug [OPTIONS] CONTAINER COMMAND [ARG...]",
		Short: "Run a command in a running container",
		Args:  RequiresMinArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.container = args[0]
			options.command = args[1:]
			return runExec(dockerCli, options)
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	flags.StringVarP(&options.image, "image", "", "", "base this image")
	flags.StringVarP(&options.detachKeys, "detach-keys", "", "", "Override the key sequence for detaching a container")
	flags.BoolVarP(&options.interactive, "interactive", "i", false, "Keep STDIN open even if not attached")
	flags.BoolVarP(&options.tty, "tty", "t", false, "Allocate a pseudo-TTY")
	flags.StringVarP(&options.user, "user", "u", "", "Username or UID (format: <name|uid>[:<group|gid>])")
	flags.BoolVarP(&options.privileged, "privileged", "", false, "Give extended privileges to the command")
	flags.StringVarP(&options.workdir, "workdir", "w", "", "Working directory inside the container")
	_ = flags.SetAnnotation("workdir", "version", []string{"1.35"})
	return cmd
}

func runExec(cli Cli, options execOptions) error {
	return nil
}
