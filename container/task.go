package container

import (
	"fmt"
	"sync"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/opencontainers/runtime-spec/specs-go"
)

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

func ExecuteCommand(containerID, taskName, cwd string, args ...string) (<-chan containerd.ExitStatus, error) {
	container, err := GetContainer(containerID)
	if err != nil {
		return nil, fmt.Errorf("ExecuteCommand: %w", err)
	}

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return nil, fmt.Errorf("ExecuteCommand: %w", err)
	}

	processSpec := new(specs.Process)
	processSpec.Terminal = true
	processSpec.Args = args
	processSpec.Cwd = cwd

	process, err := task.Exec(ctx, taskName, processSpec, cio.NewCreator(cio.WithStdio))
	if err != nil {
		if err := containerd.WithProcessKill(ctx, task); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: Exec: %w", err)
		}
		if _, err := task.Delete(ctx); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: Exec: %w", err)
		}
		return nil, fmt.Errorf("ExecuteCommand: %w", err)
	}

	if err := process.Start(ctx); err != nil {
		if err := containerd.WithProcessKill(ctx, task); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: Start: %w", err)
		}
		if _, err := task.Delete(ctx); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: Start: %w", err)
		}
		return nil, fmt.Errorf("ExecuteCommand: %w", err)
	}

	if err := SetTask(containerID, taskName, task); err != nil {
		if err := containerd.WithProcessKill(ctx, task); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: SetTask: %w", err)
		}
		if _, err := task.Delete(ctx); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: SetTask: %w", err)
		}
		return nil, fmt.Errorf("ExecuteCommand: %w", err)
	}

	exitStatusCh, err := process.Wait(ctx)
	if err != nil {
		if err := containerd.WithProcessKill(ctx, task); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: Wait: %w", err)
		}
		if _, err := task.Delete(ctx); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: Wait: %w", err)
		}
		return nil, fmt.Errorf("ExecuteCommand: %w", err)
	}

	return exitStatusCh, nil
}
