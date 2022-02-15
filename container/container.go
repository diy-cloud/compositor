package container

import (
	"fmt"
	"io"
	"sync"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/oci"
)

type ContainerMap struct {
	m map[string]containerd.Container
	sync.Mutex
}

func (c *Client) GetContainer(id string) (containerd.Container, error) {
	c.containerMap.Lock()
	container, ok := c.containerMap.m[id]
	c.containerMap.Unlock()
	if !ok {
		return nil, fmt.Errorf("GetContainer: %w", ErrNotFound)
	}
	return container, nil
}

func (c *Client) SetContainer(id string, container containerd.Container) error {
	c.containerMap.Lock()
	if _, ok := c.containerMap.m[id]; ok {
		c.containerMap.Unlock()
		return fmt.Errorf("SetContainer: %w", ErrAlreadyExists)
	}
	c.containerMap.m[id] = container
	c.containerMap.Unlock()
	return nil
}

func (c *Client) DeleteContainer(id string) error {
	container, err := c.GetContainer(id)
	if err != nil {
		return fmt.Errorf("DeleteContainer: %w", err)
	}
	if err := container.Delete(c.ctx, containerd.WithSnapshotCleanup); err != nil {
		return fmt.Errorf("DeleteContainer: %w", err)
	}
	c.containerMap.Lock()
	delete(c.containerMap.m, id)
	c.containerMap.Unlock()
	c.imageMap.Lock()
	delete(c.imageMap.m, id)
	c.imageMap.Unlock()
	return nil
}

func (c *Client) NewContainerBasedTarStream(containerID string, snapshotID string, imageReader io.Reader) error {
	if c.SnapshotExists(containerID, snapshotID) {
		return fmt.Errorf("NewContainerFromImage: Snapshot: %w", ErrAlreadyExists)
	}

	images, err := c.client.Import(c.ctx, imageReader)
	if err != nil {
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}
	image := containerd.NewImage(c.client, images[0])
	c.SetImage(image.Name(), image)

	container, err := client.NewContainer(c.ctx, containerID, containerd.WithNewSnapshot(snapshotID, image), containerd.WithNewSpec(oci.WithImageConfig(image)))
	if err != nil {
		c.DeleteImage(image.Name())
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	if err := c.SetContainer(containerID, container); err != nil {
		c.DeleteImage(image.Name())
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	if err := c.AddSnapshotToMap(containerID, snapshotID); err != nil {
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	return nil
}

func (c *Client) NewContainerBasedOnImage(containerID string, snapshotID string, imageName string) error {
	if c.SnapshotExists(containerID, snapshotID) {
		return fmt.Errorf("NewContainerFromImage: Snapshot: %w", ErrAlreadyExists)
	}

	image, err := c.GetImage(imageName)
	if err != nil {
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	container, err := client.NewContainer(c.ctx, containerID, containerd.WithNewSnapshot(snapshotID, image), containerd.WithNewSpec(oci.WithImageConfig(image)))
	if err != nil {
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	if err := c.SetContainer(containerID, container); err != nil {
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	if err := c.AddSnapshotToMap(containerID, snapshotID); err != nil {
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	return nil
}

func CopyFile() {
	// get containers.Container from container's Info
	// and get snapshottor name from snapshot's Info
	// and get snapshot from client's snapshot service
	// mount !
}
