package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type (
	// HttpServerDelegate defines the interface for a server that can listen for
	// connections and shut down gracefully.
	HttpServerDelegate interface {
		ListenAndServe() error
		Shutdown(ctx context.Context) error
	}

	// HttpServer wraps an HttpServerDelegate with graceful shutdown on OS interrupt signals.
	HttpServer struct {
		shutdown <-chan os.Signal
		delegate HttpServerDelegate
		timeout  time.Duration
	}
)

// NewHttpServer creates an HttpServer that delegates to the given server
// and listens for OS interrupt signals to trigger graceful shutdown.
func NewHttpServer(delegate HttpServerDelegate) *HttpServer {
	return &HttpServer{
		shutdown: newInterruptSignalChannel(),
		delegate: delegate,
		timeout:  time.Second * 5,
	}
}

// ListenAndServe starts the delegate server and blocks until either the server
// returns an error or an OS interrupt signal triggers graceful shutdown.
func (s *HttpServer) ListenAndServe() error {
	select {
	case err := <-s.delegateListenAndServe():
		return err
	case <-s.shutdown:
		return s.shutdownDelegate()
	}
}

// delegateListenAndServe starts the delegate in a goroutine and returns a channel
// that receives any non-ErrServerClosed error.
func (s *HttpServer) delegateListenAndServe() chan error {
	errCh := make(chan error)

	go func() {
		if err := s.delegate.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	return errCh
}

// shutdownDelegate gracefully shuts down the delegate with a timeout.
func (s *HttpServer) shutdownDelegate() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	if err := s.delegate.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		return err
	}

	return ctx.Err()
}

// newInterruptSignalChannel returns a channel that receives SIGINT, SIGHUP, SIGQUIT, and SIGTERM.
func newInterruptSignalChannel() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
	return ch
}
