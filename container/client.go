package container

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
)

var client *containerd.Client

func init() {
	var err error
	client, err = containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		panic(fmt.Errorf("container.init: %w", err))
	}
}

type Client struct {
	ctx    context.Context
	client *containerd.Client

	imageMap     ImageMap
	taskMap      TaskMap
	containerMap ContainerMap
	snapshotMap  SnapshotMap
}

var clients = []*Client{}

func NewClient(namespace string) *Client {
	c := &Client{
		ctx:    namespaces.WithNamespace(context.Background(), namespace),
		client: client,

		imageMap:     ImageMap{m: make(map[string]containerd.Image)},
		taskMap:      TaskMap{m: make(map[string]map[string]containerd.Task)},
		containerMap: ContainerMap{m: make(map[string]containerd.Container)},
		snapshotMap:  SnapshotMap{m: make(map[string]map[string]struct{})},
	}
	clients = append(clients, c)
	return c
}

func (c *Client) Close() error {
	c.taskMap.Lock()
	defer c.taskMap.Unlock()
	c.containerMap.Lock()
	defer c.containerMap.Unlock()
	c.imageMap.Lock()
	defer c.imageMap.Unlock()
	for _, containers := range c.taskMap.m {
		for _, task := range containers {
			if err := containerd.WithProcessKill(c.ctx, task); err != nil {
				return fmt.Errorf("Close: %w", err)
			}
			if _, err := task.Delete(c.ctx); err != nil {
				return fmt.Errorf("Close: %w", err)
			}
		}
	}
	for _, container := range c.containerMap.m {
		if err := container.Delete(c.ctx, containerd.WithSnapshotCleanup); err != nil {
			return fmt.Errorf("Close: %w", err)
		}
	}
	for _, image := range c.imageMap.m {
		if err := client.ImageService().Delete(c.ctx, image.Name()); err != nil {
			return fmt.Errorf("Close: DeleteImage: %w", err)
		}
	}
	return nil
}
