package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var ctx, cancel = context.WithCancel(context.Background())

var containerList = struct {
	list []string
	sync.Mutex
}{
	list: []string{},
}

func Close() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("Close: %w", err)
	}
	defer cli.Close()
	defer cancel()
	for _, c := range containerList.list {
		if err := cli.ContainerStop(ctx, c, nil); err != nil {
			return fmt.Errorf("Close: %w", err)
		}
	}
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return fmt.Errorf("Close: %w", err)
	}
	for _, i := range images {
		if _, err := cli.ImageRemove(ctx, i.ID, types.ImageRemoveOptions{
			PruneChildren: true,
		}); err != nil {
			return fmt.Errorf("Close: %w", err)
		}
	}
	return nil
}

func ImportImageFromReader(reader io.Reader, name string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("ImportImageFromReader: %w", err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp, err := cli.ImageLoad(ctx, reader, false)
	if err != nil {
		return fmt.Errorf("ImportImageFromReader: %w", err)
	}
	io.Copy(os.Stdout, resp.Body)

	return nil
}

var Ports = struct {
	Map map[int]struct{}
	sync.Mutex
}{
	Map: map[int]struct{}{},
}

const (
	PortRangeStart = 49152
	PortRangeEnd   = 65535
)

func CreateContainerByImage(image string, name string) (string, int, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", 0, fmt.Errorf("CreateContainerByImage: %w", err)
	}
	defer cli.Close()

	port := 0
	Ports.Lock()
	defer Ports.Unlock()
	for i := PortRangeStart; i <= PortRangeEnd; i++ {
		if _, ok := Ports.Map[i]; !ok {
			port = i
			Ports.Map[i] = struct{}{}
			break
		}
	}

	hostConfig := &container.HostConfig{
		AutoRemove: true,
		PortBindings: nat.PortMap{
			"80/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: strconv.Itoa(port),
				},
			},
		},
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		ExposedPorts: nat.PortSet{
			"80/tcp": struct{}{},
		},
		Tty: false,
	}, hostConfig, nil, nil, name)
	if err != nil {
		return "", 0, fmt.Errorf("CreateContainerByImage: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", 0, fmt.Errorf("CreateContainerByImage: %w", err)
	}

	containerList.Lock()
	defer containerList.Unlock()
	containerList.list = append(containerList.list, resp.ID)

	return resp.ID, port, nil
}

func RemoveContainer(id string, port int, count *int) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("RemoveContainer: %w", err)
	}

	for *count > 0 {
		runtime.Gosched()
	}

	if err := cli.ContainerStop(ctx, id, nil); err != nil {
		return fmt.Errorf("RemoveContainer: %w", err)
	}

	Ports.Lock()
	defer Ports.Unlock()
	delete(Ports.Map, port)

	return nil
}
