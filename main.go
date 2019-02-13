package main

import (
	"context"
	"fmt"

	dockertypes "github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
)

func main() {
	client, err := dockerclient.NewClient(dockerclient.DefaultDockerHost, "", nil, nil)
	if err != nil {
		panic(err)
	}
	list, err := client.ContainerList(context.Background(), dockertypes.ContainerListOptions{
		All: true,
	})
	if err != nil {
		panic(err)
	}
	for _, l := range list {
		fmt.Println(l.Names, l.Image, l.Command, l.Ports, l.Status, l.Mounts)
	}
	err = client.Close()
	if err != nil {
		panic(err)
	}
}
