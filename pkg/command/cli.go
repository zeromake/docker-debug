package command

import (
	"io"
	"os"
	"runtime"

	"github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/pkg/errors"
	"github.com/zeromake/docker-debug/pkg/cmd/version"
	"github.com/zeromake/docker-debug/pkg/opts"
	"github.com/zeromake/docker-debug/pkg/stream"
	pkgterm "github.com/zeromake/docker-debug/pkg/term"
)

type DockerCliOption func(cli *DockerCli) error

type Cli interface {
	Client() client.APIClient
	Out() *stream.OutStream
	Err() io.Writer
	In() *stream.InStream
	SetIn(in *stream.InStream)
	ServerInfo() ServerInfo
	ClientInfo() ClientInfo
}

type DockerCli struct {
	in         *stream.InStream
	out        *stream.OutStream
	err        io.Writer
	client     client.APIClient
	serverInfo ServerInfo
	clientInfo ClientInfo
}

func NewDockerCli(ops ...DockerCliOption) (*DockerCli, error) {
	cli := &DockerCli{}
	if err := cli.Apply(ops...); err != nil {
		return nil, err
	}
	if cli.out == nil || cli.in == nil || cli.err == nil {
		stdin, stdout, stderr := pkgterm.StdStreams()
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
func (cli *DockerCli) Apply(ops ...DockerCliOption) error {
	for _, op := range ops {
		if err := op(cli); err != nil {
			return err
		}
	}
	return nil
}

func getServerHost(hosts []string, tlsOptions *tlsconfig.Options) (string, error) {
	var host string
	switch len(hosts) {
	case 0:
		host = os.Getenv("DOCKER_HOST")
	case 1:
		host = hosts[0]
	default:
		return "", errors.New("Please specify only one -H")
	}

	return opts.ParseHost(tlsOptions != nil, host)
}

// UserAgent returns the user agent string used for making API requests
func UserAgent() string {
	return "Docker-Debug-Client/" + version.Version + " (" + runtime.GOOS + ")"
}

// ServerInfo stores details about the supported features and platform of the
// server
type ServerInfo struct {
	HasExperimental bool
	OSType          string
	BuildkitVersion string
}

// ClientInfo stores details about the supported features of the client
type ClientInfo struct {
	HasExperimental bool
	DefaultVersion  string
}

// DefaultVersion returns api.defaultVersion or DOCKER_API_VERSION if specified.
func (cli *DockerCli) DefaultVersion() string {
	return cli.clientInfo.DefaultVersion
}

// Client returns the APIClient
func (cli *DockerCli) Client() client.APIClient {
	return cli.client
}

// Out returns the writer used for stdout
func (cli *DockerCli) Out() *stream.OutStream {
	return cli.out
}

// Err returns the writer used for stderr
func (cli *DockerCli) Err() io.Writer {
	return cli.err
}

// SetIn sets the reader used for stdin
func (cli *DockerCli) SetIn(in *stream.InStream) {
	cli.in = in
}

// In returns the reader used for stdin
func (cli *DockerCli) In() *stream.InStream {
	return cli.in
}

// ServerInfo returns the server version details for the host this client is
// connected to
func (cli *DockerCli) ServerInfo() ServerInfo {
	return cli.serverInfo
}

// ClientInfo returns the client details for the cli
func (cli *DockerCli) ClientInfo() ClientInfo {
	return cli.clientInfo
}
