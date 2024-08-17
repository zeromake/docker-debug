package command

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	dockerImage "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/zeromake/docker-debug/internal/config"
	"github.com/zeromake/docker-debug/pkg/opts"
	"github.com/zeromake/docker-debug/pkg/stream"
	"github.com/zeromake/docker-debug/pkg/tty"
)

const (
	caKey   = "ca.pem"
	certKey = "cert.pem"
	keyKey  = "key.pem"

	legacyDefaultDomain = "index.docker.io"
	defaultDomain       = "docker.io"
	officialRepoName    = "library"
)

// DebugCliOption cli option
type DebugCliOption func(cli *DebugCli) error

// Cli interface
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

// DebugCli cli struct
type DebugCli struct {
	in     *stream.InStream
	out    *stream.OutStream
	err    io.Writer
	client client.APIClient
	config *config.Config
	ctx    context.Context
}

// NewDebugCli new DebugCli
func NewDebugCli(ctx context.Context, ops ...DebugCliOption) (*DebugCli, error) {
	cli := &DebugCli{ctx: ctx}
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

// Apply all the operation on the cli
func (cli *DebugCli) Apply(ops ...DebugCliOption) error {
	for _, op := range ops {
		if err := op(cli); err != nil {
			return err
		}
	}
	return nil
}

// WithConfig set config
func WithConfig(config *config.Config) DebugCliOption {
	return func(cli *DebugCli) error {
		cli.config = config
		return nil
	}
}

// WithClientConfig set docker config
func WithClientConfig(dockerConfig *config.DockerConfig) DebugCliOption {
	return func(cli *DebugCli) error {
		if cli.client != nil {
			err := cli.client.Close()
			if err != nil {
				return errors.WithStack(err)
			}
		}
		var (
			host string
			err  error
		)
		host, err = opts.ValidateHost(dockerConfig.Host)
		if err != nil {
			return err
		}
		clientOpts := []client.Opt{
			client.WithHost(host),
			client.WithVersion(dockerConfig.Version),
		}
		if dockerConfig.TLS {
			clientOpts = append(clientOpts, client.WithTLSClientConfig(
				fmt.Sprintf("%s%s%s", dockerConfig.CertDir, config.PathSeparator, caKey),
				fmt.Sprintf("%s%s%s", dockerConfig.CertDir, config.PathSeparator, certKey),
				fmt.Sprintf("%s%s%s", dockerConfig.CertDir, config.PathSeparator, keyKey),
			))
		}
		dockerClient, err := client.NewClientWithOpts(clientOpts...)
		if err != nil {
			return errors.WithStack(err)
		}
		cli.client = dockerClient
		return nil
	}
}

// UserAgent returns the user agent string used for making API requests

// Close cli close
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

// Config config
func (cli *DebugCli) Config() *config.Config {
	return cli.config
}

// splitDockerDomain splits a repository name to domain and remotename string.
// If no valid domain is found, the default domain is used. Repository name
// needs to be already validated before.
func splitDockerDomain(name string) (domain, remainder string) {
	i := strings.IndexRune(name, '/')
	if i == -1 || (!strings.ContainsAny(name[:i], ".:") && name[:i] != "localhost") {
		domain, remainder = defaultDomain, name
	} else {
		domain, remainder = name[:i], name[i+1:]
	}
	if domain == legacyDefaultDomain {
		domain = defaultDomain
	}
	if domain == defaultDomain && !strings.ContainsRune(remainder, '/') {
		remainder = officialRepoName + "/" + remainder
	}
	return
}

// PullImage pull docker image
func (cli *DebugCli) PullImage(image string) error {
	domain, remainder := splitDockerDomain(image)
	imageName := path.Join(domain, remainder)

	ctx, cancel := context.WithCancel(cli.ctx)
	defer cancel()
	responseBody, err := cli.client.ImagePull(ctx, imageName, dockerImage.PullOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		err = responseBody.Close()
		if err != nil {
			logrus.Debugf("%+v", err)
		}
	}()
	return jsonmessage.DisplayJSONMessagesToStream(responseBody, cli.out, nil)
}

// FindImage find image
func (cli *DebugCli) FindImage(image string) ([]dockerImage.Summary, error) {
	args := filters.NewArgs()
	args.Add("reference", image)
	ctx, cancel := cli.withContent(cli.config.Timeout)
	defer cancel()
	return cli.client.ImageList(ctx, dockerImage.ListOptions{
		Filters: args,
	})
}

// Ping ping docker
func (cli *DebugCli) Ping() (types.Ping, error) {
	ctx, cancel := cli.withContent(cli.config.Timeout)
	defer cancel()
	return cli.client.Ping(ctx)
}

func (cli *DebugCli) withContent(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(cli.ctx, timeout)
}

func containerMode(name string) string {
	return fmt.Sprintf("container:%s", name)
}

// CreateContainer create new container and attach target container resource
func (cli *DebugCli) CreateContainer(attachContainer string, options execOptions) (string, error) {
	var mounts []mount.Mount
	ctx, cancel := cli.withContent(cli.config.Timeout)
	info, err := cli.client.ContainerInspect(ctx, attachContainer)
	cancel()
	if err != nil {
		return "", errors.WithStack(err)
	}
	if !info.State.Running {
		return "", errors.Errorf("container: `%s` is not running", attachContainer)
	}
	attachContainer = info.ID
	mergedDir, ok := info.GraphDriver.Data["MergedDir"]
	if !ok || mergedDir == "" {
		return "", fmt.Errorf("container: `%s` not found merged dir", attachContainer)
	}
	if cli.config.MountDir != "" {
		mounts = append(mounts, mount.Mount{
			Type:   "bind",
			Source: mergedDir,
			Target: cli.config.MountDir,
		})
		for _, i := range info.Mounts {
			var mountType = i.Type
			if i.Type == "volume" {
				mountType = "bind"
			}
			mounts = append(mounts, mount.Mount{
				Type:     mountType,
				Source:   i.Source,
				Target:   cli.config.MountDir + i.Destination,
				ReadOnly: !i.RW,
			})
		}
	}
	if options.volumes != nil {
		// -v bind mount
		if mounts == nil {
			mounts = []mount.Mount{}
		}
		for _, m := range options.volumes {
			mountArgs := strings.Split(m, ":")
			mountLen := len(mountArgs)
			if mountLen > 0 && mountLen <= 3 {
				if strings.HasPrefix(mountArgs[0], "$c/") {
					mountArgs[0] = path.Join(mergedDir, mountArgs[0][11:])
				}
				mountDefault := mount.Mount{
					Type:     "bind",
					ReadOnly: false,
				}
				switch mountLen {
				case 1:
					mountDefault.Source = mountArgs[0]
					mountDefault.Target = mountArgs[0]
				case 2:
					if mountArgs[1] == "rw" || mountArgs[1] == "ro" {
						mountDefault.ReadOnly = mountArgs[1] != "rw"
						mountDefault.Source = mountArgs[0]
						mountDefault.Target = mountArgs[0]
					} else {
						mountDefault.Source = mountArgs[0]
						mountDefault.Target = mountArgs[1]
					}
				case 3:
					mountDefault.Source = mountArgs[0]
					mountDefault.Target = mountArgs[1]
					mountDefault.ReadOnly = mountArgs[2] != "rw"
				}
				mounts = append(mounts, mountDefault)
			}
		}
	}
	targetName := containerMode(attachContainer)

	conf := &container.Config{
		Entrypoint: strslice.StrSlice([]string{"/usr/bin/env", "sh"}),
		Image:      cli.config.Image,
		Tty:        true,
		OpenStdin:  true,
		StdinOnce:  true,
		StopSignal: "SIGKILL",
	}
	hostConfig := &container.HostConfig{
		NetworkMode: container.NetworkMode(targetName),
		UsernsMode:  container.UsernsMode(":" + attachContainer),
		PidMode:     container.PidMode(targetName),
		Mounts:      mounts,
		SecurityOpt: options.securityOpts,
		CapAdd:      options.capAdds,
		AutoRemove:  true,
		Privileged:  options.privileged,
	}

	// default is not use ipc
	if options.ipc {
		hostConfig.IpcMode = container.IpcMode(targetName)
	}
	ctx, cancel = cli.withContent(cli.config.Timeout)
	body, err := cli.client.ContainerCreate(
		ctx,
		conf,
		hostConfig,
		nil,
		nil,
		"",
	)
	cancel()
	if err != nil {
		return "", errors.WithStack(err)
	}
	ctx, cancel = cli.withContent(cli.config.Timeout)
	err = cli.client.ContainerStart(
		ctx,
		body.ID,
		container.StartOptions{},
	)
	cancel()
	return body.ID, errors.WithStack(err)
}

// ContainerClean stop and remove container
func (cli *DebugCli) ContainerClean(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	var timeout int = 5
	return errors.WithStack(cli.client.ContainerStop(
		ctx,
		id,
		container.StopOptions{Timeout: &timeout},
	))
}

// ExecCreate exec create
func (cli *DebugCli) ExecCreate(options execOptions, containerStr string) (types.IDResponse, error) {
	var workDir = options.workDir
	if workDir == "" && cli.config.MountDir != "" {
		workDir = path.Join(cli.config.MountDir, options.targetDir)
	}
	h, w := cli.out.GetTtySize()
	opt := container.ExecOptions{
		User:         options.user,
		Privileged:   options.privileged,
		DetachKeys:   options.detachKeys,
		Tty:          true,
		AttachStderr: true,
		AttachStdin:  true,
		AttachStdout: true,
		WorkingDir:   workDir,
		Cmd:          options.command,
		ConsoleSize:  &[2]uint{h, w},
	}
	ctx, cancel := cli.withContent(cli.config.Timeout)
	defer cancel()
	resp, err := cli.client.ContainerExecCreate(ctx, containerStr, opt)
	return resp, errors.WithStack(err)
}

// ExecStart exec start
func (cli *DebugCli) ExecStart(options execOptions, execID string) error {
	h, w := cli.out.GetTtySize()
	execConfig := container.ExecStartOptions{
		Tty:         true,
		ConsoleSize: &[2]uint{h, w},
	}

	ctx, cancel := cli.withContent(cli.config.Timeout)
	defer cancel()
	response, err := cli.client.ContainerExecAttach(ctx, execID, execConfig)
	if err != nil {
		return errors.WithStack(err)
	}
	defer response.Close()
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		streamer := tty.HijackedIOStreamer{
			Streams:      cli,
			InputStream:  cli.in,
			OutputStream: cli.out,
			ErrorStream:  cli.err,
			Resp:         response,
			TTY:          true,
			DetachKeys:   options.detachKeys,
		}
		errCh <- streamer.Stream(cli.ctx)
	}()
	if err := tty.MonitorTtySize(cli.ctx, cli.client, cli.out, execID, true); err != nil {
		_, _ = fmt.Fprintln(cli.err, "Error monitoring TTY size:", err)
	}
	if err := <-errCh; err != nil {
		logrus.Debugf("Error hijack: %s", err)
		return err
	}
	return getExecExitStatus(cli.ctx, cli.client, execID)
}

// WatchContainer watch container
func (cli *DebugCli) WatchContainer(ctx context.Context, containerID string) error {
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	filterArgs := filters.NewArgs()
	filterArgs.Add("container", containerID)
	messages, errs := cli.client.Events(subCtx, events.ListOptions{
		Filters: filterArgs,
	})

	for {
		select {
		case event := <-messages:
			if event.Type == events.ContainerEventType {
				switch event.Action {
				case events.ActionDestroy, events.ActionDie, events.ActionKill, events.ActionStop:
					return nil
				}
			}
		case err := <-errs:
			return err
		}
	}
}

func getExecExitStatus(ctx context.Context, apiClient client.ContainerAPIClient, execID string) error {
	resp, err := apiClient.ContainerExecInspect(ctx, execID)
	if err != nil {
		// If we can't connect, then the daemon probably died.
		if !client.IsErrConnectionFailed(err) {
			return err
		}
		return errors.Errorf("ExitStatus %d", -1)
	}
	status := resp.ExitCode
	if status != 0 {
		return errors.Errorf("ExitStatus %d", status)
	}
	return nil
}

// FindContainer find container
func (cli *DebugCli) FindContainer(name string) (string, error) {
	return name, nil
}
