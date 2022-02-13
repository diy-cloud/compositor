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

func GetContainer(id string) (containerd.Container, error) {
	containerMap.Lock()
	container, ok := containerMap.m[id]
	containerMap.Unlock()
	if !ok {
		return nil, fmt.Errorf("GetContainer: %w", ErrNotFound)
	}
	return container, nil
}

func SetContainer(id string, container containerd.Container) error {
	containerMap.Lock()
	if _, ok := containerMap.m[id]; ok {
		containerMap.Unlock()
		return fmt.Errorf("SetContainer: %w", ErrAlreadyExists)
	}
	containerMap.m[id] = container
	containerMap.Unlock()
	return nil
}

func DeleteContainer(id string) error {
	container, err := GetContainer(id)
	if err != nil {
		return fmt.Errorf("DeleteContainer: %w", err)
	}
	if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		return fmt.Errorf("DeleteContainer: %w", err)
	}
	containerMap.Lock()
	delete(containerMap.m, id)
	containerMap.Unlock()
	imageMap.Lock()
	delete(imageMap.m, id)
	imageMap.Unlock()
	return nil
}

var containerMap = ContainerMap{
	m: make(map[string]containerd.Container),
}

func NewContainerBasedTarStream(containerID string, snapshotID string, imageReader io.Reader) error {
	if SnapshotExists(containerID, snapshotID) {
		return fmt.Errorf("NewContainerFromImage: Snapshot: %w", ErrAlreadyExists)
	}

	images, err := client.Import(ctx, imageReader)
	if err != nil {
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}
	image := containerd.NewImage(client, images[0])
	SetImage(image.Name(), image)

	container, err := client.NewContainer(ctx, containerID, containerd.WithNewSnapshot(snapshotID, image), containerd.WithNewSpec(oci.WithImageConfig(image)))
	if err != nil {
		DeleteImage(image.Name())
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	if err := SetContainer(containerID, container); err != nil {
		DeleteImage(image.Name())
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	if err := AddSnapshotToMap(containerID, snapshotID); err != nil {
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	return nil
}

func NewContainerBasedOnImage(containerID string, snapshotID string, imageName string) error {
	if SnapshotExists(containerID, snapshotID) {
		return fmt.Errorf("NewContainerFromImage: Snapshot: %w", ErrAlreadyExists)
	}

	image, err := GetImage(imageName)
	if err != nil {
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	container, err := client.NewContainer(ctx, containerID, containerd.WithNewSnapshot(snapshotID, image), containerd.WithNewSpec(oci.WithImageConfig(image)))
	if err != nil {
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	if err := SetContainer(containerID, container); err != nil {
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	if err := AddSnapshotToMap(containerID, snapshotID); err != nil {
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	return nil
}
