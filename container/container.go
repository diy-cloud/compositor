package container

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
)

var ctx = namespaces.WithNamespace(context.Background(), "compositor")

var client *containerd.Client

type ContainerMap struct {
	m map[string]containerd.Container
	sync.Mutex
}

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")

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

type ImageMap struct {
	m map[string]containerd.Image
	sync.Mutex
}

func GetImage(id string) (containerd.Image, error) {
	imageMap.Lock()
	image, ok := imageMap.m[id]
	imageMap.Unlock()
	if !ok {
		return nil, fmt.Errorf("GetImage: %w", ErrNotFound)
	}
	return image, nil
}

func SetImage(id string, image containerd.Image) error {
	imageMap.Lock()
	if _, ok := imageMap.m[id]; ok {
		imageMap.Unlock()
		return ErrAlreadyExists
	}
	imageMap.m[id] = image
	imageMap.Unlock()
	return nil
}

func DeleteImage(id string) error {
	if err := client.ImageService().Delete(ctx, id); err != nil {
		return fmt.Errorf("DeleteImage: %w", err)
	}
	imageMap.Lock()
	delete(imageMap.m, id)
	imageMap.Unlock()
	return nil
}

var imageMap = ImageMap{
	m: make(map[string]containerd.Image),
}

type TaskMap struct {
	m map[string]map[string]containerd.Task
	sync.Mutex
}

func GetTask(containerID string, taskName string) (containerd.Task, error) {
	taskMap.Lock()
	tasks, ok := taskMap.m[containerID]
	taskMap.Unlock()
	if !ok {
		return nil, fmt.Errorf("GetTask: %w", ErrNotFound)
	}
	task, ok := tasks[taskName]
	if !ok {
		return nil, fmt.Errorf("GetTask: %w", ErrNotFound)
	}
	return task, nil
}

func SetTask(containerID string, taskName string, task containerd.Task) error {
	taskMap.Lock()
	if _, ok := taskMap.m[containerID]; !ok {
		taskMap.m[containerID] = make(map[string]containerd.Task)
	}
	if _, ok := taskMap.m[containerID][taskName]; ok {
		taskMap.Unlock()
		return fmt.Errorf("SetTask: %w", ErrAlreadyExists)
	}
	taskMap.m[containerID][taskName] = task
	taskMap.Unlock()
	return nil
}

func DeleteTask(containerID string, taskName string) error {
	task, err := GetTask(containerID, taskName)
	if err != nil {
		return fmt.Errorf("DeleteTask: %w", err)
	}
	if err := containerd.WithProcessKill(ctx, task); err != nil {
		return fmt.Errorf("DeleteTask: %w", err)
	}
	if _, err := task.Delete(ctx); err != nil {
		return fmt.Errorf("DeleteTask: %w", err)
	}
	taskMap.Lock()
	delete(taskMap.m[containerID], taskName)
	taskMap.Unlock()
	return nil
}

var taskMap = TaskMap{
	m: make(map[string]map[string]containerd.Task),
}

type SnapshotMap struct {
	m map[string]map[string]struct{}
	sync.Mutex
}

func SnapshotExists(containerID, snapshotID string) bool {
	snapshotMap.Lock()
	snapshots, ok := snapshotMap.m[containerID]
	if !ok {
		return false
	}
	if _, ok := snapshots[snapshotID]; !ok {
		return false
	}
	snapshotMap.Unlock()
	return true
}

func AddSnapshotToMap(containerID, snapshotID string) error {
	snapshotMap.Lock()
	if _, ok := snapshotMap.m[containerID]; !ok {
		snapshotMap.m[containerID] = make(map[string]struct{})
	}
	if _, ok := snapshotMap.m[containerID][snapshotID]; ok {
		snapshotMap.Unlock()
		return fmt.Errorf("AddSnapshotToMap: %w", ErrAlreadyExists)
	}
	snapshotMap.m[containerID][snapshotID] = struct{}{}
	snapshotMap.Unlock()
	return nil
}

func DeleteSnapshotFromMap(containerID, snapshotID string) error {
	snapshotMap.Lock()
	if _, ok := snapshotMap.m[containerID]; !ok {
		snapshotMap.Unlock()
		return fmt.Errorf("DeleteSnapshotFromMap: %w", ErrNotFound)
	}
	if _, ok := snapshotMap.m[containerID][snapshotID]; !ok {
		snapshotMap.Unlock()
		return fmt.Errorf("DeleteSnapshotFromMap: %w", ErrNotFound)
	}
	delete(snapshotMap.m[containerID], snapshotID)
	snapshotMap.Unlock()
	return nil
}

var snapshotMap = SnapshotMap{
	m: make(map[string]map[string]struct{}),
}

func init() {
	var err error
	client, err = containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		panic(fmt.Errorf("container.init: %w", err))
	}
}

func NewImagesFromTarStream(tar io.Reader) ([]string, error) {
	images, err := client.Import(ctx, tar)
	if err != nil {
		return nil, fmt.Errorf("NewImageFromTar: %w", err)
	}
	names := make([]string, len(images))
	for i := 0; i < len(images); i++ {
		image := containerd.NewImage(client, images[0])
		imageMap.Lock()
		imageMap.m[image.Name()] = image
		imageMap.Unlock()
		names[i] = image.Name()
	}
	return names, nil
}

func NewImageFromURL(url string) (string, error) {
	image, err := client.Pull(ctx, url, containerd.WithPullUnpack)
	if err != nil {
		return "", fmt.Errorf("NewImageFromURL: %w", err)
	}
	imageMap.Lock()
	imageMap.m[image.Name()] = image
	imageMap.Unlock()
	return image.Name(), nil
}

func NewContainerFromImage(containerID string, snapshotID string, imageReader io.Reader) error {
	if SnapshotExists(containerID, snapshotID) {
		return fmt.Errorf("NewContainerFromImage: Snapshot: %w", ErrAlreadyExists)
	}

	images, err := client.Import(ctx, imageReader)
	if err != nil {
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}
	image := containerd.NewImage(client, images[0])
	SetImage(image.Name(), image)

	container, err := client.NewContainer(ctx, containerID, containerd.WithNewSnapshot(snapshotID, image), containerd.WithImage(image))
	if err != nil {
		DeleteImage(image.Name())
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}

	if err := SetContainer(containerID, container); err != nil {
		DeleteImage(image.Name())
		return fmt.Errorf("NewContainerFromImage: %w", err)
	}
	return nil
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
		if err := container.Delete(ctx); err != nil {
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
