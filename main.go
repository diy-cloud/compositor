package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	networkResp, err := cli.NetworkCreate(ctx, "test", types.NetworkCreate{
		Driver:         "bridge",
		CheckDuplicate: true,
		Scope:          "local",
		Internal:       true,
	})
	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(ctx, "docker.io/yeasy/simple-web", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "yeasy/simple-web",
		Tty:   false,
	}, nil, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"network": {
				NetworkID: networkResp.ID,
			},
		},
	}, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case data := <-statusCh:
		fmt.Println(data.StatusCode)
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			scanner := bufio.NewReader(out)
			line, _, err := scanner.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}
			fmt.Printf("[%s-%s]: %s\n", "alpine", resp.ID[:10], string(line[8:]))
		}
	}()

	terminalChan := make(chan os.Signal, 1)
	signal.Notify(terminalChan)
	<-terminalChan

	if err := cli.ContainerStop(ctx, resp.ID, nil); err != nil {
		panic(err)
	}

	if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}); err != nil {
		panic(err)
	}

	list, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("container length: %v\n", len(list))
	for _, container := range list {
		fmt.Println(container.ID)
	}

	if err := cli.NetworkRemove(ctx, networkResp.ID); err != nil {
		panic(err)
	}
}
