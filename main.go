package main

import (
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/snowmerak/compositor/proxy"
	"golang.org/x/net/http2"
)

func main() {
	go func() {
		server := &http.Server{
			Addr:           "0.0.0.0:9999",
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
