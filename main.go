package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Server is a container for web server context.  Implements http.Server
// interface.
type Server struct {
	StaticFileServer http.Handler // http.FileServer  used to server static files out of ./assets/static
}

// notFound dishes out 404 responses.
func notFound(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>404 - %q is a bogus url sucka.</h1>", r.URL.Path)
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// simple url dispatcher but first, we should strip any trailing slashes from
	// the path.
	if url := r.URL.Path; len(url) > 1 && url[len(url)-1] == '/' {
		http.Redirect(w, r, url[:len(url)-1], http.StatusMovedPermanently)
		return
	}

	// dispatch urls to apropriate handlers.
	switch {
	case strings.HasPrefix(r.URL.Path, "/static/"):
		s.StaticFileServer.ServeHTTP(w, r)

	case strings.HasPrefix(r.URL.Path, "/b/"):
		w.Write([]byte("blog post"))

	case r.URL.Path == "/":
		w.Write([]byte("hello, world!"))

	default:
		notFound(w, r)
	}
}

func sig() <-chan os.Signal {
	rc := make(chan os.Signal)
	signal.Notify(rc, syscall.SIGINT)

	return rc
}

func main() {
	svr := http.Server{
		Handler: Server{
			StaticFileServer: http.StripPrefix("/static/", http.FileServer(http.Dir("./assets"))),
		},
		Addr:              ":8080",
		ReadHeaderTimeout: 5 * time.Second,
	}

	errchan := make(chan error)

	go func() { errchan <- svr.ListenAndServe() }()

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
