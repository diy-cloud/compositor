package main

import (
	"fmt"
	"log"

	"github.com/snowmerak/compositor/container"
)

func main() {
	end := container.End()

	imageName, err := container.NewImageFromURL("docker.io/library/alpine:latest")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(imageName)

	if err := container.NewContainerBasedOnImage("container12", "snapshot12", imageName); err != nil {
		log.Println(err)
	}
	fmt.Println("made container4")

	exitStatus, err := container.ExecuteCommand("container12", "echo", "/", "/bin/sh", "-c", "echo HelloWorld!")
	if err != nil {
		log.Println(err)
	}
	fmt.Println("executed echo")

	if exitStatus != nil {
		status := <-exitStatus
		fmt.Println(status.Result())
	}

	<-end
}
