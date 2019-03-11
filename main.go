package main

import (
	"net/http"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
	"crypto/tls"
	"github.com/pkg/errors"

	"github.com/docker/docker/api/types"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerfilters "github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	term "github.com/zeromake/docker-debug/utils"
	"crypto/x509"
	"encoding/pem"
)

const imageName = "nicolaka/netshoot:latest"

const containerName = "project_redis_*"

var command = []string{
	"bash",
}
const (
	defaultCertDir = "C:\\Users\\Administrator\\.docker\\machine\\certs"
	caKey   = "ca.pem"
	certKey = "cert.pem"
	keyKey  = "key.pem"
)
var clientCipherSuites = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
}

// TLSData tls 配置
type TLSData struct {
	CA   []byte
	Key  []byte
	Cert []byte
}
// TLSDataFromFiles 从证书文件加载tls配置
func TLSDataFromFiles(caPath, certPath, keyPath string) (*TLSData, error) {
	var (
		ca, cert, key []byte
		err           error
	)
	if caPath != "" {
		if ca, err = ioutil.ReadFile(caPath); err != nil {
			return nil, err
		}
	}
	if certPath != "" {
		if cert, err = ioutil.ReadFile(certPath); err != nil {
			return nil, err
		}
	}
	if keyPath != "" {
		if key, err = ioutil.ReadFile(keyPath); err != nil {
			return nil, err
		}
	}
	if ca == nil && cert == nil && key == nil {
		return nil, nil
	}
	return &TLSData{CA: ca, Cert: cert, Key: key}, nil
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
			_, _ = io.Copy(resp.Conn, inputStream)
		}
		_ = resp.CloseWrite()
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
func tlsConfig(tlsData *TLSData) (*tls.Config, error) {
	if tlsData == nil {
		// there is no specific tls config
		return nil, nil
	}
	tlsconfig := &tls.Config{
		// Prefer TLS1.2 as the client minimum
		MinVersion:   tls.VersionTLS12,
		CipherSuites: clientCipherSuites,
	}
	if tlsData != nil && tlsData.CA != nil {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(tlsData.CA) {
			return nil, errors.New("failed to retrieve context tls info: ca.pem seems invalid")
		}
		tlsconfig.RootCAs = certPool
	}
	if tlsData != nil && tlsData.Key != nil && tlsData.Cert != nil {
		keyBytes := tlsData.Key
		pemBlock, _ := pem.Decode(keyBytes)
		if pemBlock == nil {
			return nil, fmt.Errorf("no valid private key found")
		}

		var err error
		if x509.IsEncryptedPEMBlock(pemBlock) {
			keyBytes, err = x509.DecryptPEMBlock(pemBlock, []byte(""))
			if err != nil {
				return nil, errors.Wrap(err, "private key is encrypted, but could not decrypt it")
			}
			keyBytes = pem.EncodeToMemory(&pem.Block{Type: pemBlock.Type, Bytes: keyBytes})
		}

		x509cert, err := tls.X509KeyPair(tlsData.Cert, keyBytes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve context tls info")
		}
		tlsconfig.Certificates = []tls.Certificate{x509cert}
	}
	// if c.SkipTLSVerify {
	// 	tlsOpts = append(tlsOpts, func(cfg *tls.Config) {
	// 		cfg.InsecureSkipVerify = true
	// 	})
	// }

	return tlsconfig, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tlsData, err := TLSDataFromFiles(
		fmt.Sprintf("%s\\%s", defaultCertDir, caKey),
		fmt.Sprintf("%s\\%s", defaultCertDir, certKey),
		fmt.Sprintf("%s\\%s", defaultCertDir, keyKey),
	)
	if err != nil {
		panic(err)
	}
	tlsconfig, err := tlsConfig(tlsData)
	if err != nil {
		panic(err)
	}
	httpTransport := &http.Transport{
		TLSClientConfig: tlsconfig,
	}
	httpClient := &http.Client{
		Transport: httpTransport,
	}
	client, err := dockerclient.NewClient("tcp://192.168.99.100:2376", "", httpClient, map[string]string{
		"User-Agent": "docker-debug-v0.1.0",
	})

	defer func() {
		// err = client.Close()
		// if err != nil {
		// 	panic(err)
		// }
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
		_ = term.DisplayJSONMessagesStream(out, os.Stdout, 1, true, nil)
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
		info, err := client.ContainerInspect(ctx, containerObj.ID)

		if err != nil {
			panic(err)
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
		// for _, i := range info.Mounts {
		// 	mounts = append(mounts, mount.Mount{
		// 		Type:     i.Type,
		// 		Source:   i.Source,
		// 		Target:   "/mnt/container" + i.Destination,
		// 		ReadOnly: !i.RW,
		// 	})
		// }
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
			// VolumesFrom: []string{
			// 	targetID,
			// },
			Mounts: mounts,
		}
		body, err := client.ContainerCreate(ctx, config, hostConfig, nil, "")
		if err != nil {
			panic(err)
		}
		defer func() {
			var timeout = time.Second * 3
			_ = client.ContainerStop(ctx, body.ID, &timeout)
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
