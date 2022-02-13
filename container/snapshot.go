package container

import (
	"fmt"
	"sync"
)

type SnapshotMap struct {
	m map[string]map[string]struct{}
	sync.Mutex
}

func SnapshotExists(containerID, snapshotID string) bool {
	snapshotMap.Lock()
	defer snapshotMap.Unlock()
	snapshots, ok := snapshotMap.m[containerID]
	if !ok {
		return false
	}
	if _, ok := snapshots[snapshotID]; !ok {
		return false
	}
	return true
}

func AddSnapshotToMap(containerID, snapshotID string) error {
	snapshotMap.Lock()
	defer snapshotMap.Unlock()
	if _, ok := snapshotMap.m[containerID]; !ok {
		snapshotMap.m[containerID] = make(map[string]struct{})
	}
	if _, ok := snapshotMap.m[containerID][snapshotID]; ok {
		snapshotMap.Unlock()
		return fmt.Errorf("AddSnapshotToMap: %w", ErrAlreadyExists)
	}
	snapshotMap.m[containerID][snapshotID] = struct{}{}
	return nil
}

func DeleteSnapshotFromMap(containerID, snapshotID string) error {
	snapshotMap.Lock()
	defer snapshotMap.Unlock()
	if _, ok := snapshotMap.m[containerID]; !ok {
		return fmt.Errorf("DeleteSnapshotFromMap: %w", ErrNotFound)
	}
	if _, ok := snapshotMap.m[containerID][snapshotID]; !ok {
		return fmt.Errorf("DeleteSnapshotFromMap: %w", ErrNotFound)
	}
	delete(snapshotMap.m[containerID], snapshotID)
	return nil
}

var snapshotMap = SnapshotMap{
	m: make(map[string]map[string]struct{}),
}
