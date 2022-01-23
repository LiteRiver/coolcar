package main

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func main() {
	c, err := client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	resp, err := c.ContainerCreate(
		ctx,
		&container.Config{
			Image: "mongo:latest",
			ExposedPorts: nat.PortSet{
				"27017": {},
			},
		}, &container.HostConfig{
			PortBindings: nat.PortMap{
				"27017": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "0",
					},
				},
			},
		},
		nil,
		nil,
		"",
	)

	if err != nil {
		panic(err)
	}

	err = c.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println("container started")
	time.Sleep(10 * time.Second)

	insp, err := c.ContainerInspect(ctx, resp.ID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("mongo listening at: %+v\n", insp.NetworkSettings.Ports["27017/tcp"][0].HostPort)
	fmt.Println("removing container")
	err = c.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{
		Force: true,
	})

	if err != nil {
		panic(err)
	}
}
