package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct{}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HELLO!"))
}

func sig() <-chan os.Signal {
	rc := make(chan os.Signal)
	signal.Notify(rc, syscall.SIGHUP, syscall.SIGINT)

	return rc
}

func main() {

	svr := http.Server{
		Handler:           Server{},
		Addr:              ":8080",
		ReadHeaderTimeout: 5 * time.Second,
	}

	errchan := make(chan error)

	go func() {
		errchan <- svr.ListenAndServe()
	}()

	sigint := sig()

	for {
		select {
		case err := <-errchan:
			log.Fatal(err)

		case s := <-sigint:
			log.Printf("received %s; shutting down server.", s)
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// shudown and give server 5 seconds to settle down.
				svr.Shutdown(ctx)

				// once server has shut down, svr.ListenAndServe() will have returned
				// an error value and reading from the errchan should handle properly
				// exiting.
			}()
		}
	}
}
