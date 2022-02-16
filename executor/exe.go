package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/snowmerak/compositor/docker"
	"github.com/snowmerak/compositor/proxy"
	"github.com/snowmerak/compositor/router/register"
	"github.com/snowmerak/lux"
	"golang.org/x/net/http2"
)

func main() {
	go func() {
		app := lux.New(nil)
		registerGroup := app.NewRouterGroup("/register")
		registerGroup.POST("/:id", register.Post, nil)

		if err := app.ListenAndServe2TLS(":9999", "localhost/cert.pem", "localhost/key.pem"); err != nil {
			panic(err)
		}
	}()

	go func() {
		server := &http.Server{
			Addr:           "0.0.0.0:80",
			Handler:        http.HandlerFunc(proxy.Handler),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		http2.ConfigureServer(server, nil)

		fmt.Println("External HTTP Server is Listening on :80")
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	go func() {
		server := &http.Server{
			Addr:           "0.0.0.0:443",
			Handler:        http.HandlerFunc(proxy.Handler),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		http2.ConfigureServer(server, nil)

		fmt.Println("External HTTPS Server is Listening on :443")
		if err := server.ListenAndServeTLS("localhost/cert.pem", "localhost/key.pem"); err != nil {
			panic(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Signal(syscall.SIGTERM))
	<-ch

	if err := docker.Close(); err != nil {
		log.Println(err)
	}
}
