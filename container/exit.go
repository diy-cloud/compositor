package container

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

var terminalSignal = make(chan os.Signal, 1)

var end = make(chan struct{}, 1)

func init() {
	signal.Notify(terminalSignal, os.Interrupt, os.Signal(syscall.SIGTERM))

	go func() {
		<-terminalSignal
		for _, c := range clients {
			if err := c.Close(); err != nil {
				log.Println(err)
			}
		}
		if err := client.Close(); err != nil {
			log.Println(err)
		}
		end <- struct{}{}
		signal.Stop(terminalSignal)
		signal.Reset(os.Interrupt, os.Signal(syscall.SIGTERM))
		close(terminalSignal)
	}()
}

func End() <-chan struct{} {
	return end
}
