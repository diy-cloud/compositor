package container

import (
	"fmt"
	"io"
	"sync"

	"github.com/containerd/containerd"
)

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
