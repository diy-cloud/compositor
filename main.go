package main

import (
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/snowmerak/compositor/proxy"
	configRouter "github.com/snowmerak/compositor/router/config"
	"github.com/snowmerak/compositor/router/register"
	"github.com/snowmerak/lux"
	"golang.org/x/net/http2"
)

func main() {
	go func() {
		app := lux.New(nil)

		registerGroup := app.NewRouterGroup("/register")
		registerGroup.GET("/:name", register.Get, nil)

		configGroup := app.NewRouterGroup("/config")
		configGroup.GET("/:name", configRouter.Get, nil)

		if err := app.ListenAndServe2TLS("0.0.0.0:8080", "localhost/cert.pem", "localhost/key.pem"); err != nil {
			panic(err)
		}
	}()

	go func() {
		server := &http.Server{
			Addr:           "0.0.0.0:8080",
			Handler:        http.HandlerFunc(proxy.Handler),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		http2.ConfigureServer(server, nil)

		if err := server.ListenAndServeTLS("localhost/cert.pem", "localhost/key.pem"); err != nil {
			panic(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	go func() {
		signal.Notify(ch, os.Interrupt)
	}()
	<-ch
}
