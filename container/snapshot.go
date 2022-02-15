package container

import (
	"fmt"
	"sync"
)

type SnapshotMap struct {
	m map[string]map[string]struct{}
	sync.Mutex
}

func (c *Client) SnapshotExists(containerID, snapshotID string) bool {
	c.snapshotMap.Lock()
	defer c.snapshotMap.Unlock()
	snapshots, ok := c.snapshotMap.m[containerID]
	if !ok {
		return false
	}
	if _, ok := snapshots[snapshotID]; !ok {
		return false
	}
	return true
}

func (c *Client) AddSnapshotToMap(containerID, snapshotID string) error {
	c.snapshotMap.Lock()
	defer c.snapshotMap.Unlock()
	if _, ok := c.snapshotMap.m[containerID]; !ok {
		c.snapshotMap.m[containerID] = make(map[string]struct{})
	}
	if _, ok := c.snapshotMap.m[containerID][snapshotID]; ok {
		c.snapshotMap.Unlock()
		return fmt.Errorf("AddSnapshotToMap: %w", ErrAlreadyExists)
	}
	c.snapshotMap.m[containerID][snapshotID] = struct{}{}
	return nil
}

func (c *Client) DeleteSnapshotFromMap(containerID, snapshotID string) error {
	c.snapshotMap.Lock()
	defer c.snapshotMap.Unlock()
	if _, ok := c.snapshotMap.m[containerID]; !ok {
		return fmt.Errorf("DeleteSnapshotFromMap: %w", ErrNotFound)
	}
	if _, ok := c.snapshotMap.m[containerID][snapshotID]; !ok {
		return fmt.Errorf("DeleteSnapshotFromMap: %w", ErrNotFound)
	}
	delete(c.snapshotMap.m[containerID], snapshotID)
	return nil
}
