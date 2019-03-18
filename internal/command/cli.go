package command

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zeromake/docker-debug/internal/config"
	"github.com/zeromake/docker-debug/pkg/tty"
	"github.com/zeromake/moby/api/types"
	"github.com/zeromake/moby/api/types/container"
	"github.com/zeromake/moby/api/types/filters"
	"github.com/zeromake/moby/api/types/mount"
	"github.com/zeromake/moby/api/types/strslice"
	"github.com/zeromake/moby/pkg/jsonmessage"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/zeromake/docker-debug/cmd/version"
	"github.com/zeromake/docker-debug/pkg/stream"
	"github.com/zeromake/moby/client"
	"github.com/zeromake/moby/pkg/term"
)

const (
	caKey   = "ca.pem"
	certKey = "cert.pem"
	keyKey  = "key.pem"
)

type DebugCliOption func(cli *DebugCli) error

type Cli interface {
	Client() client.APIClient
	Out() *stream.OutStream
	Err() io.Writer
	In() *stream.InStream
	SetIn(in *stream.InStream)
	PullImage(image string) error
	FindImage(image string) error
	Config() *config.Config
}

type DebugCli struct {
	in     *stream.InStream
	out    *stream.OutStream
	err    io.Writer
	client client.APIClient
	config *config.Config
}

func NewDebugCli(ops ...DebugCliOption) (*DebugCli, error) {
	cli := &DebugCli{}
	if err := cli.Apply(ops...); err != nil {
		return nil, err
	}
	if cli.out == nil || cli.in == nil || cli.err == nil {
		stdin, stdout, stderr := term.StdStreams()
		if cli.in == nil {
			cli.in = stream.NewInStream(stdin)
		}
		if cli.out == nil {
			cli.out = stream.NewOutStream(stdout)
		}
		if cli.err == nil {
			cli.err = stderr
		}
	}
	return cli, nil
}

func NewDefaultDebugCli() (*DebugCli, error) {
	return NewDebugCli(WithConfigFile(), WithClientName("default"))
}

// Apply all the operation on the cli
func (cli *DebugCli) Apply(ops ...DebugCliOption) error {
	for _, op := range ops {
		if err := op(cli); err != nil {
			return err
		}
	}
	return nil
}

func WithConfigFile() DebugCliOption {
	return func(cli *DebugCli) error {
		conf, err := config.LoadConfig()
		if err != nil {
			return errors.WithStack(err)
		}
		cli.config = conf
		return nil
	}
}

func WithConfig(config *config.Config) DebugCliOption {
	return func(cli *DebugCli) error {
		cli.config = config
		return nil
	}
}

func WithClientConfig(dockerConfig config.DockerConfig) DebugCliOption {
	return func(cli *DebugCli) error {
		if cli.client != nil {
			err := cli.client.Close()
			if err != nil {
				return errors.WithStack(err)
			}
		}
		opts := []func(*client.Client) error{
			client.WithHost(dockerConfig.Host),
			client.WithVersion(""),
		}
		if dockerConfig.TLS {
			opts = append(opts, client.WithTLSClientConfig(
				fmt.Sprintf("%s%s%s", dockerConfig.CertDir, config.PathSeparator, caKey),
				fmt.Sprintf("%s%s%s", dockerConfig.CertDir, config.PathSeparator, certKey),
				fmt.Sprintf("%s%s%s", dockerConfig.CertDir, config.PathSeparator, keyKey),
			))
		}
		dockerClient, err := client.NewClientWithOpts(opts...)
		if err != nil {
			return errors.WithStack(err)
		}
		cli.client = dockerClient
		return nil
	}
}

func WithClientName(name string) DebugCliOption {
	return func(cli *DebugCli) error {
		dockerConfig := cli.config.DockerConfig[name]
		return WithClientConfig(dockerConfig)(cli)
	}
}

// UserAgent returns the user agent string used for making API requests
func UserAgent() string {
	return "Docker-Debug-Client/" + version.Version + " (" + runtime.GOOS + ")"
}

func (cli *DebugCli) Close() error {
	if cli.client != nil {
		return errors.WithStack(cli.client.Close())
	}
	return nil
}

// Client returns the APIClient
func (cli *DebugCli) Client() client.APIClient {
	return cli.client
}

// Out returns the writer used for stdout
func (cli *DebugCli) Out() *stream.OutStream {
	return cli.out
}

// Err returns the writer used for stderr
func (cli *DebugCli) Err() io.Writer {
	return cli.err
}

// SetIn sets the reader used for stdin
func (cli *DebugCli) SetIn(in *stream.InStream) {
	cli.in = in
}

// In returns the reader used for stdin
func (cli *DebugCli) In() *stream.InStream {
	return cli.in
}

func (cli *DebugCli) Config() *config.Config {
	return cli.config
}

func (cli *DebugCli) PullImage(image string) error {
	var name = image
	temps := strings.SplitN(name, "/", 1)
	if strings.IndexRune(temps[0], '.') == -1 {
		name = "docker.io/" + image
	}
	responseBody, err := cli.client.ImagePull(cli.withContent(cli.config.Timeout*30), name, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	defer responseBody.Close()
	return jsonmessage.DisplayJSONMessagesToStream(responseBody, cli.out, nil)
}

func (cli *DebugCli) FindImage(image string) ([]types.ImageSummary, error) {
	args := filters.NewArgs()
	args.Add("reference", image)
	return cli.client.ImageList(cli.withContent(cli.config.Timeout), types.ImageListOptions{
		Filters: args,
	})
}

func (cli *DebugCli) Ping() (types.Ping, error) {
	return cli.client.Ping(cli.withContent(cli.config.Timeout))
}

func (cli *DebugCli) withContent(timeout time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	return ctx
}

func containerMode(name string) string {
	return fmt.Sprintf("container:%s", name)
}

func (cli *DebugCli) CreateContainer(attachContainer string) (string, error) {
	info, err := cli.client.ContainerInspect(cli.withContent(cli.config.Timeout), attachContainer)
	if err != nil {
		return "", errors.WithStack(err)
	}
	mountDir, ok := info.GraphDriver.Data["MergedDir"]
	mounts := []mount.Mount{}
	if ok {
		mounts = append(mounts, mount.Mount{
			Type:   "bind",
			Source: mountDir,
			Target: "/mnt/container",
		})
	}
	for _, i := range info.Mounts {
		if i.Type == "volume" {
			continue
		}
		mounts = append(mounts, mount.Mount{
			Type:     i.Type,
			Source:   i.Source,
			Target:   "/mnt/container" + i.Destination,
			ReadOnly: !i.RW,
		})
	}

	targetName := containerMode(attachContainer)

	conf := &container.Config{
		Entrypoint: strslice.StrSlice([]string{"/usr/bin/env", "sh"}),
		Image:      cli.config.Image,
		Tty:        true,
		OpenStdin:  true,
		StdinOnce:  true,
	}
	hostConfig := &container.HostConfig{
		NetworkMode: container.NetworkMode(targetName),
		UsernsMode:  container.UsernsMode(targetName),
		IpcMode:     container.IpcMode(targetName),
		PidMode:     container.PidMode(targetName),
		Mounts:      mounts,
	}
	body, err := cli.client.ContainerCreate(
		cli.withContent(cli.config.Timeout),
		conf,
		hostConfig,
		nil,
		"",
	)
	if err != nil {
		return "", errors.WithStack(err)
	}
	err = cli.client.ContainerStart(
		cli.withContent(cli.config.Timeout),
		body.ID,
		types.ContainerStartOptions{},
	)
	return body.ID, errors.WithStack(err)
}

func (cli *DebugCli) ContainerClean(id string) error {
	//_ = cli.client.ContainerStop(
	//	cli.withContent(time.Second * 2),
	//	id,
	//	&cli.config.Timeout,
	//)

	return errors.WithStack(cli.client.ContainerRemove(
		cli.withContent(cli.config.Timeout),
		id,
		types.ContainerRemoveOptions{
			Force: true,
		},
	))
}

func (cli *DebugCli) ExecCreate(options execOptions, container string) (types.IDResponse, error) {
	opt := types.ExecConfig{
		User:         options.user,
		Privileged:   options.privileged,
		DetachKeys:   options.detachKeys,
		Tty:          true,
		AttachStderr: true,
		AttachStdin:  true,
		AttachStdout: true,
		WorkingDir:   options.workdir,
		Cmd:          options.command,
	}
	resp, err := cli.client.ContainerExecCreate(cli.withContent(cli.config.Timeout), container, opt)
	return resp, errors.WithStack(err)
}

func (cli *DebugCli) ExecStart(options execOptions, execID string) error {
	execConfig := types.ExecStartCheck{
		Tty:          true,
	}

	response, err := cli.client.ContainerExecAttach(cli.withContent(cli.config.Timeout), execID, execConfig)
	if err != nil {
		return errors.WithStack(err)
	}
	streamer := tty.HijackedIOStreamer{
		Streams:      cli,
		InputStream:  cli.in,
		OutputStream: cli.out,
		ErrorStream:  cli.err,
		Resp:         response,
		TTY:          true,
	}
	return streamer.Stream(context.Background());
}

func (cli *DebugCli) FindContainer(name string) (string, error) {
	containerArgs := filters.NewArgs()
	containerArgs.Add("name", name)
	list, err := cli.client.ContainerList(cli.withContent(cli.config.Timeout), types.ContainerListOptions{
		Filters: containerArgs,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}
	listLen := len(list)
	if listLen == 1 {
		return list[0].ID, nil
	}
	if listLen == 0 {
		return "", errors.Errorf("not find %s container!", name)
	}
	var containerNames = []string{}
	for _, c := range list {
		containerNames = append(containerNames, strings.Join(c.Names, "-"))
	}
	return "", errors.Errorf("ContainerList:\n%s\n", strings.Join(containerNames, "\n"))
}
