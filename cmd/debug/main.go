package main

import "github.com/zeromake/docker-debug/internal/command"

func main() {
	command.Execute()
}
//
//import (
//	"context"
//	"errors"
//	"fmt"
//	"net/http"
//	"os"
//	"strings"
//	"time"
//
//	"github.com/zeromake/docker-debug/pkg/stream"
//	"github.com/zeromake/docker-debug/pkg/tty"
//	"github.com/zeromake/moby/api/types"
//	dockertypes "github.com/zeromake/moby/api/types"
//	"github.com/zeromake/moby/api/types/container"
//	dockerfilters "github.com/zeromake/moby/api/types/filters"
//	"github.com/zeromake/moby/api/types/mount"
//	"github.com/zeromake/moby/api/types/strslice"
//	dockerclient "github.com/zeromake/moby/client"
//	"github.com/zeromake/moby/pkg/jsonmessage"
//	"github.com/zeromake/moby/pkg/signal"
//	"github.com/zeromake/moby/pkg/term"
//)
//
//const imageName = "nicolaka/netshoot:latest"
//
//const containerName = "caishichang_mysql_*"
//
//var command = []string{
//	"bash",
//}
//
//// const (
//// 	defaultCertDir = "C:\\Users\\Administrator\\.docker\\machine\\certs"
//// 	caKey          = "ca.pem"
//// 	certKey        = "cert.pem"
//// 	keyKey         = "key.pem"
//// )
//
//type streams struct {
//	out *stream.OutStream
//	in  *stream.InStream
//}
//
//func (s *streams) Out() *stream.OutStream {
//	return s.out
//}
//func (s *streams) In() *stream.InStream {
//	return s.in
//}
//
//func containerMode(name string) string {
//	return fmt.Sprintf("container:%s", name)
//}
//
//// func holdHijackedConnection(tty bool, inputStream io.Reader, outputStream, errorStream io.Writer, resp dockertypes.HijackedResponse) error {
//// 	receiveStdout := make(chan error)
//// 	if outputStream != nil || errorStream != nil {
//// 		go func() {
//// 			receiveStdout <- redirectResponseToOutputStream(tty, outputStream, errorStream, resp.Reader)
//// 		}()
//// 	}
//
//// 	stdinDone := make(chan struct{})
//// 	go func() {
//// 		if inputStream != nil {
//// 			_, _ = io.Copy(resp.Conn, inputStream)
//// 		}
//// 		_ = resp.CloseWrite()
//// 		close(stdinDone)
//// 	}()
//
//// 	select {
//// 	case err := <-receiveStdout:
//// 		return err
//// 	case <-stdinDone:
//// 		if outputStream != nil || errorStream != nil {
//// 			return <-receiveStdout
//// 		}
//// 	}
//// 	return nil
//// }
//
//// func redirectResponseToOutputStream(tty bool, outputStream, errorStream io.Writer, resp io.Reader) error {
//// 	if outputStream == nil {
//// 		outputStream = ioutil.Discard
//// 	}
//// 	if errorStream == nil {
//// 		errorStream = ioutil.Discard
//// 	}
//// 	var err error
//// 	if tty {
//// 		_, err = io.Copy(outputStream, resp)
//// 	} else {
//// 		_, err = stdcopy.StdCopy(outputStream, errorStream, resp)
//// 	}
//// 	return err
//// }
//func resizeTTY(ctx context.Context, out *stream.OutStream, client dockerclient.ContainerAPIClient, containerID string) {
//	height, width := out.GetTtySize()
//	tty.ResizeTtyTo(
//		ctx,
//		client,
//		containerID,
//		height+1,
//		width+1,
//		false,
//	)
//	if err := tty.MonitorTtySize(
//		ctx,
//		client,
//		out,
//		containerID,
//		false,
//	); err != nil {
//		fmt.Printf("Error monitoring TTY size: %s\n", err)
//	}
//}
//
//func main() {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	stdin, stdout, stderr := term.StdStreams()
//	// tlsconfig, err := debug.TLSConfigFromFiles(
//	// 	fmt.Sprintf("%s\\%s", defaultCertDir, caKey),
//	// 	fmt.Sprintf("%s\\%s", defaultCertDir, certKey),
//	// 	fmt.Sprintf("%s\\%s", defaultCertDir, keyKey),
//	// 	"",
//	// 	false,
//	// )
//	// if err != nil {
//	// 	panic(err)
//	// }
//	// httpTransport := &http.Transport{
//	// 	TLSClientConfig: tlsconfig,
//	// }
//	// httpClient := &http.Client{
//	// 	Transport: httpTransport,
//	// }
//	var httpClient *http.Client = nil
//	// "tcp://192.168.99.100:2376"
//	client, err := dockerclient.NewClient(dockerclient.DefaultDockerHost, "", httpClient, map[string]string{
//		"User-Agent": "docker-debug-v0.1.0",
//	})
//	if err != nil {
//		panic(err)
//	}
//	defer func() {
//		err = client.Close()
//		if err != nil {
//			panic(err)
//		}
//	}()
//	out := stream.NewOutStream(stdout)
//	args := dockerfilters.NewArgs()
//	args.Add("reference", imageName)
//	images, err := client.ImageList(ctx, dockertypes.ImageListOptions{
//		Filters: args,
//	})
//	if err != nil {
//		panic(err)
//	}
//	if len(images) == 0 {
//		var name = imageName
//		temps := strings.SplitN(name, "/", 1)
//		if strings.IndexRune(temps[0], '.') == -1 {
//			name = "docker.io/" + imageName
//		}
//		responseBody, err := client.ImagePull(ctx, name, dockertypes.ImagePullOptions{})
//		if err != nil {
//			panic(err)
//		}
//		err = jsonmessage.DisplayJSONMessagesToStream(responseBody, out, nil)
//		// _ = term.DisplayJSONMessagesStream(responseBody, os.Stdout, 1, true, nil)
//		responseBody.Close()
//		if err != nil {
//			panic(err)
//		}
//	} else {
//		fmt.Printf("Image: %s is has\n", imageName)
//	}
//
//	// https://docs.docker.com/engine/api/v1.39/#operation/ContainerList
//	containerArgs := dockerfilters.NewArgs()
//	if strings.IndexRune(containerName, '*') != -1 {
//		containerArgs.Add("name", containerName)
//	} else {
//		containerArgs.Add("since", containerName)
//	}
//	containerArgs.Add("status", "running")
//	containerList, err := client.ContainerList(ctx, dockertypes.ContainerListOptions{
//		Filters: containerArgs,
//	})
//	if err != nil {
//		panic(err)
//	}
//
//	if len(containerList) == 1 {
//		containerObj := containerList[0]
//		info, err := client.ContainerInspect(ctx, containerObj.ID)
//
//		if err != nil {
//			panic(err)
//		}
//		mountDir, ok := info.GraphDriver.Data["MergedDir"]
//		mounts := []mount.Mount{}
//		if ok {
//			mounts = append(mounts, mount.Mount{
//				Type:   "bind",
//				Source: mountDir,
//				Target: "/mnt/container",
//			})
//		}
//		// for _, i := range info.Mounts {
//		// 	mounts = append(mounts, mount.Mount{
//		// 		Type:     i.Type,
//		// 		Source:   i.Source,
//		// 		Target:   "/mnt/container" + i.Destination,
//		// 		ReadOnly: !i.RW,
//		// 	})
//		// }
//		config := &container.Config{
//			Entrypoint: strslice.StrSlice(command),
//			Image:      imageName,
//			Tty:        true,
//			OpenStdin:  true,
//			StdinOnce:  true,
//		}
//		var targetID = containerObj.ID
//		var targetName = containerMode(targetID)
//		hostConfig := &container.HostConfig{
//			NetworkMode: container.NetworkMode(targetName),
//			UsernsMode:  container.UsernsMode(targetName),
//			IpcMode:     container.IpcMode(targetName),
//			PidMode:     container.PidMode(targetName),
//			// VolumesFrom: []string{
//			// 	targetID,
//			// },
//			Mounts: mounts,
//		}
//		body, err := client.ContainerCreate(ctx, config, hostConfig, nil, "")
//		if err != nil {
//			panic(err)
//		}
//		defer func() {
//			var timeout = time.Second * 3
//			_ = client.ContainerStop(ctx, body.ID, &timeout)
//			err = client.ContainerRemove(ctx, body.ID, types.ContainerRemoveOptions{})
//			if err != nil {
//				panic(err)
//			}
//		}()
//		err = client.ContainerStart(ctx, body.ID, types.ContainerStartOptions{})
//		if err != nil {
//			panic(err)
//		}
//		var TTY = true
//
//		// opts := dockertypes.ContainerAttachOptions{
//		// 	Stream: true,
//		// 	Stdin:  true,
//		// 	Stdout: true,
//		// 	Stderr: true,
//		// }
//		if !TTY {
//			sigc := tty.ForwardAllSignals(ctx, client, body.ID)
//			defer signal.StopCatch(sigc)
//		}
//		execConfig := dockertypes.ExecConfig{
//			AttachStdin:  true,
//			AttachStderr: true,
//			AttachStdout: true,
//			Tty:          true,
//			Cmd:          command,
//		}
//
//		response, err := client.ContainerExecCreate(ctx, body.ID, execConfig)
//		if err != nil {
//			panic(err)
//		}
//
//		execID := response.ID
//		if execID == "" {
//			panic(errors.New("exec ID empty"))
//		}
//		execStartCheck := dockertypes.ExecStartCheck{
//			Tty: true,
//		}
//		resp, err := client.ContainerExecAttach(ctx, execID, execStartCheck)
//		if err != nil {
//			panic(err)
//		}
//		if TTY && out.IsTerminal() {
//			resizeTTY(ctx, out, client, body.ID)
//		}
//		in := stream.NewInStream(stdin)
//		s := &streams{
//			out: out,
//			in:  in,
//		}
//		if !TTY {
//			stderr = os.Stdout
//		}
//		streamer := tty.HijackedIOStreamer{
//			Streams:      s,
//			InputStream:  in,
//			OutputStream: out,
//			ErrorStream:  stderr,
//			Resp:         resp,
//			TTY:          TTY,
//		}
//		// go func() {
//		// 	timer := time.NewTimer(time.Second * 2)
//		// 	<-timer.C
//		// 	_, _ = os.Stdin.WriteString("ls\n")
//		// }()
//		if err := streamer.Stream(ctx); err != nil {
//			panic(err)
//		}
//
//		// err = holdHijackedConnection(true, os.Stdin, os.Stdout, os.Stderr, resp)
//		// if err != nil {
//		// 	panic(err)
//		// }
//	} else {
//		for _, containerItem := range containerList {
//			fmt.Printf("Container:\t%s\t%s\n", containerItem.Names[0], containerItem.ID)
//		}
//	}
//}
