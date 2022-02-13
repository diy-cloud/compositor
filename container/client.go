package container

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
)

var ctx = namespaces.WithNamespace(context.Background(), "compositor")

var client *containerd.Client

func init() {
	var err error
	client, err = containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		panic(fmt.Errorf("container.init: %w", err))
	}
}

func Close() error {
	taskMap.Lock()
	defer taskMap.Unlock()
	containerMap.Lock()
	defer containerMap.Unlock()
	imageMap.Lock()
	defer imageMap.Unlock()
	for _, containers := range taskMap.m {
		for _, task := range containers {
			if err := containerd.WithProcessKill(ctx, task); err != nil {
				return fmt.Errorf("Close: %w", err)
			}
			if _, err := task.Delete(ctx); err != nil {
				return fmt.Errorf("Close: %w", err)
			}
		}
	}
	for _, container := range containerMap.m {
		if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
			return fmt.Errorf("Close: %w", err)
		}
	}
	for _, image := range imageMap.m {
		if err := client.ImageService().Delete(ctx, image.Name()); err != nil {
			return fmt.Errorf("Close: DeleteImage: %w", err)
		}
	}
	if err := client.Close(); err != nil {
		return fmt.Errorf("Close: %w", err)
	}
	return nil
}
