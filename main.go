package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerfilters "github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/strslice"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	term "github.com/zeromake/docker-debug/utils"
)

const imageName = "nicolaka/netshoot:latest"

const containerName = "gaodingx_mysql_1"

var command = []string{
	"bash",
}

func containerMode(name string) string {
	return fmt.Sprintf("container:%s", name)
}
func holdHijackedConnection(tty bool, inputStream io.Reader, outputStream, errorStream io.Writer, resp dockertypes.HijackedResponse) error {
	receiveStdout := make(chan error)
	if outputStream != nil || errorStream != nil {
		go func() {
			receiveStdout <- redirectResponseToOutputStream(tty, outputStream, errorStream, resp.Reader)
		}()
	}

	stdinDone := make(chan struct{})
	go func() {
		if inputStream != nil {
			io.Copy(resp.Conn, inputStream)
		}
		resp.CloseWrite()
		close(stdinDone)
	}()

	select {
	case err := <-receiveStdout:
		return err
	case <-stdinDone:
		if outputStream != nil || errorStream != nil {
			return <-receiveStdout
		}
	}
	return nil
}

func redirectResponseToOutputStream(tty bool, outputStream, errorStream io.Writer, resp io.Reader) error {
	if outputStream == nil {
		outputStream = ioutil.Discard
	}
	if errorStream == nil {
		errorStream = ioutil.Discard
	}
	var err error
	if tty {
		_, err = io.Copy(outputStream, resp)
	} else {
		_, err = stdcopy.StdCopy(outputStream, errorStream, resp)
	}
	return err
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := dockerclient.NewClient("https://192.168.99.100:2376", "18.09.0", nil, map[string]string{
		"User-Agent": "docker-debug-v0.1.0",
	})

	defer func() {
		err = client.Close()
		if err != nil {
			panic(err)
		}
	}()
	if err != nil {
		panic(err)
	}
	args := dockerfilters.NewArgs()
	args.Add("reference", imageName)
	images, err := client.ImageList(ctx, dockertypes.ImageListOptions{
		Filters: args,
	})
	if err != nil {
		panic(err)
	}
	if len(images) == 0 {
		var name = imageName
		temps := strings.SplitN(name, "/", 1)
		if strings.IndexRune(temps[0], '.') == -1 {
			name = "docker.io/" + imageName
		}
		out, err := client.ImagePull(ctx, name, dockertypes.ImagePullOptions{})
		if err != nil {
			panic(err)
		}
		term.DisplayJSONMessagesStream(out, os.Stdout, 1, true, nil)
		out.Close()
	} else {
		fmt.Printf("Image: %s is has\n", imageName)
	}

	// https://docs.docker.com/engine/api/v1.39/#operation/ContainerList
	containerArgs := dockerfilters.NewArgs()
	if strings.IndexRune(containerName, '*') != -1 {
		containerArgs.Add("name", containerName)
	} else {
		containerArgs.Add("since", containerName)
	}
	containerArgs.Add("status", "running")
	containerList, err := client.ContainerList(ctx, dockertypes.ContainerListOptions{
		Filters: containerArgs,
	})
	if err != nil {
		panic(err)
	}

	if len(containerList) == 1 {
		containerObj := containerList[0]
		config := &container.Config{
			Entrypoint: strslice.StrSlice(command),
			Image:      imageName,
			Tty:        true,
			OpenStdin:  true,
			StdinOnce:  true,
		}
		var targetID = containerObj.ID
		var targetName = containerMode(targetID)
		hostConfig := &container.HostConfig{
			NetworkMode: container.NetworkMode(targetName),
			UsernsMode:  container.UsernsMode(targetName),
			IpcMode:     container.IpcMode(targetName),
			PidMode:     container.PidMode(targetName),
			VolumesFrom: []string{
				targetID,
			},
		}
		body, err := client.ContainerCreate(ctx, config, hostConfig, nil, "")
		if err != nil {
			panic(err)
		}
		defer func() {
			var timeout = time.Second * 3
			client.ContainerStop(ctx, body.ID, &timeout)
			err = client.ContainerRemove(ctx, body.ID, types.ContainerRemoveOptions{})
			if err != nil {
				panic(err)
			}
		}()
		err = client.ContainerStart(ctx, body.ID, types.ContainerStartOptions{})
		if err != nil {
			panic(err)
		}

		opts := dockertypes.ContainerAttachOptions{
			Stream: true,
			Stdin:  true,
			Stdout: true,
			Stderr: true,
		}

		resp, err := client.ContainerAttach(ctx, body.ID, opts)
		if err != nil {
			panic(err)
		}
		err = holdHijackedConnection(true, os.Stdin, os.Stdout, os.Stderr, resp)
		if err != nil {
			panic(err)
		}
	} else {
		for _, containerItem := range containerList {
			fmt.Printf("Container:\t%s\t%s\n", containerItem.Names[0], containerItem.ID)
		}
	}
}
