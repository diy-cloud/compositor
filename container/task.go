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

func (c *Client) GetTask(containerID string, taskName string) (containerd.Task, error) {
	c.taskMap.Lock()
	tasks, ok := c.taskMap.m[containerID]
	c.taskMap.Unlock()
	if !ok {
		return nil, fmt.Errorf("GetTask: %w", ErrNotFound)
	}
	task, ok := tasks[taskName]
	if !ok {
		return nil, fmt.Errorf("GetTask: %w", ErrNotFound)
	}
	return task, nil
}

func (c *Client) SetTask(containerID string, taskName string, task containerd.Task) error {
	c.taskMap.Lock()
	if _, ok := c.taskMap.m[containerID]; !ok {
		c.taskMap.m[containerID] = make(map[string]containerd.Task)
	}
	if _, ok := c.taskMap.m[containerID][taskName]; ok {
		c.taskMap.Unlock()
		return fmt.Errorf("SetTask: %w", ErrAlreadyExists)
	}
	c.taskMap.m[containerID][taskName] = task
	c.taskMap.Unlock()
	return nil
}

func (c *Client) DeleteTask(containerID string, taskName string) error {
	task, err := c.GetTask(containerID, taskName)
	if err != nil {
		return fmt.Errorf("DeleteTask: %w", err)
	}
	if err := containerd.WithProcessKill(c.ctx, task); err != nil {
		return fmt.Errorf("DeleteTask: %w", err)
	}
	if _, err := task.Delete(c.ctx); err != nil {
		return fmt.Errorf("DeleteTask: %w", err)
	}
	c.taskMap.Lock()
	delete(c.taskMap.m[containerID], taskName)
	c.taskMap.Unlock()
	return nil
}

func (c *Client) ExecuteCommand(containerID, taskName, cwd string, args ...string) (<-chan containerd.ExitStatus, error) {
	container, err := c.GetContainer(containerID)
	if err != nil {
		return nil, fmt.Errorf("ExecuteCommand: %w", err)
	}

	task, err := container.NewTask(c.ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return nil, fmt.Errorf("ExecuteCommand: %w", err)
	}

	processSpec := new(specs.Process)
	processSpec.Terminal = true
	processSpec.Args = args
	processSpec.Cwd = cwd

	process, err := task.Exec(c.ctx, taskName, processSpec, cio.NewCreator(cio.WithStdio))
	if err != nil {
		if err := containerd.WithProcessKill(c.ctx, task); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: Exec: %w", err)
		}
		if _, err := task.Delete(c.ctx); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: Exec: %w", err)
		}
		return nil, fmt.Errorf("ExecuteCommand: %w", err)
	}

	if err := process.Start(c.ctx); err != nil {
		if err := containerd.WithProcessKill(c.ctx, task); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: Start: %w", err)
		}
		if _, err := task.Delete(c.ctx); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: Start: %w", err)
		}
		return nil, fmt.Errorf("ExecuteCommand: %w", err)
	}

	if err := c.SetTask(containerID, taskName, task); err != nil {
		if err := containerd.WithProcessKill(c.ctx, task); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: SetTask: %w", err)
		}
		if _, err := task.Delete(c.ctx); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: SetTask: %w", err)
		}
		return nil, fmt.Errorf("ExecuteCommand: %w", err)
	}

	exitStatusCh, err := process.Wait(c.ctx)
	if err != nil {
		if err := containerd.WithProcessKill(c.ctx, task); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: Wait: %w", err)
		}
		if _, err := task.Delete(c.ctx); err != nil {
			return nil, fmt.Errorf("ExecuteCommand: Wait: %w", err)
		}
		return nil, fmt.Errorf("ExecuteCommand: %w", err)
	}

	return exitStatusCh, nil
}
