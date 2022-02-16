package register

import (
	"fmt"
	"runtime"
	"time"

	"github.com/snowmerak/compositor/docker"
	"github.com/snowmerak/compositor/proxy"
	"github.com/snowmerak/lux/context"
)

func Post(lc *context.LuxContext) error {
	bodyReader := lc.GetBodyReader()

	name := lc.GetPathVariable("id")

	containerName, workCount, oldPort, err := proxy.RemoveProxyServer(name)
	if err != nil {
		lc.SetInternalServerError()
		return fmt.Errorf("Register.Post: %w", err)
	}

	if containerName != "" && workCount != nil {
		waitCount := 1
		for {
			if err := docker.RemoveContainer(containerName, oldPort, workCount); err != nil {
				time.Sleep(time.Second * time.Duration(waitCount))
				waitCount++
				runtime.Gosched()
				continue
			}
			break
		}
	}

	if err := docker.ImportImageFromReader(bodyReader, name); err != nil {
		lc.SetBadRequest()
		return fmt.Errorf("Register.Post: %w", err)
	}

	id, newPort, err := docker.CreateContainerByImage(name, name)
	if err != nil {
		lc.SetInternalServerError()
		return fmt.Errorf("Register.Post: %w", err)
	}

	if err := proxy.AddProxyServer(name, id, newPort); err != nil {
		lc.SetInternalServerError()
		return fmt.Errorf("Register.Post: %w", err)
	}

	lc.SetOK()
	return lc.ReplyString("enabled")
}
