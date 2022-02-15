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

func (c *Client) GetImage(id string) (containerd.Image, error) {
	c.imageMap.Lock()
	image, ok := c.imageMap.m[id]
	c.imageMap.Unlock()
	if !ok {
		return nil, fmt.Errorf("GetImage: %w", ErrNotFound)
	}
	return image, nil
}

func (c *Client) SetImage(id string, image containerd.Image) error {
	c.imageMap.Lock()
	if _, ok := c.imageMap.m[id]; ok {
		c.imageMap.Unlock()
		return ErrAlreadyExists
	}
	c.imageMap.m[id] = image
	c.imageMap.Unlock()
	return nil
}

func (c *Client) DeleteImage(id string) error {
	if err := client.ImageService().Delete(c.ctx, id); err != nil {
		return fmt.Errorf("DeleteImage: %w", err)
	}
	c.imageMap.Lock()
	delete(c.imageMap.m, id)
	c.imageMap.Unlock()
	return nil
}

func (c *Client) NewImagesFromTarStream(tar io.Reader) ([]string, error) {
	images, err := client.Import(c.ctx, tar)
	if err != nil {
		return nil, fmt.Errorf("NewImageFromTar: %w", err)
	}
	names := make([]string, len(images))
	for i := 0; i < len(images); i++ {
		image := containerd.NewImage(client, images[0])
		c.imageMap.Lock()
		c.imageMap.m[image.Name()] = image
		c.imageMap.Unlock()
		names[i] = image.Name()
	}
	return names, nil
}

func (c *Client) NewImageFromURL(url string) (string, error) {
	image, err := client.Pull(c.ctx, url, containerd.WithPullUnpack)
	if err != nil {
		return "", fmt.Errorf("NewImageFromURL: %w", err)
	}
	c.imageMap.Lock()
	c.imageMap.m[image.Name()] = image
	c.imageMap.Unlock()
	return image.Name(), nil
}
