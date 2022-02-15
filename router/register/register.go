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

	if err := docker.ImportImageFromReader(bodyReader, name); err != nil {
		lc.SetBadRequest()
		return fmt.Errorf("Register.Post: %w", err)
	}

	id, port, err := docker.CreateContainerByImage(name, name)
	if err != nil {
		lc.SetInternalServerError()
		return fmt.Errorf("Register.Post: %w", err)
	}

	containerName, workCount, port, err := proxy.RemoveProxyServer(name)
	if err != nil {
		lc.SetInternalServerError()
		return fmt.Errorf("Register.Post: %w", err)
	}

	if containerName != "" && workCount != nil {
		go func() {
			waitCount := 1
			for {
				if err := docker.RemoveContainer(containerName, port, workCount); err != nil {
					time.Sleep(time.Second * time.Duration(waitCount))
					waitCount++
					runtime.Gosched()
					continue
				}
				break
			}
		}()
	}

	if err := proxy.AddProxyServer(name, id, port); err != nil {
		lc.SetInternalServerError()
		return fmt.Errorf("Register.Post: %w", err)
	}

	lc.SetOK()
	return lc.ReplyString("enabled")
}
