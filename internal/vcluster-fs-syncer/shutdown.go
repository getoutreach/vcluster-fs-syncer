// Code managed by Bootstrap.
//
// Please edit this to more accurately match the server implementation.

package vcluster_fs_syncer //nolint:revive
import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/getoutreach/gobox/pkg/orerr"
)

type ShutdownService struct {
	done chan struct{}
}

func NewShutdownService() *ShutdownService {
	return &ShutdownService{
		done: make(chan struct{}),
	}
}

func (s *ShutdownService) Run(ctx context.Context, _ *Config) error {
	// listen for interrupts and gracefully shutdown server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	select {
	case out := <-c:
		// Allow interrupt signals to be caught again in worse-case scenario
		// situations when the service hangs during a graceful shutdown.
		signal.Reset(os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

		err := fmt.Errorf("shutting down due to interrupt: %v", out)
		return orerr.ShutdownError{Err: err}
	case <-ctx.Done():
		return ctx.Err()
	case <-s.done:
		return nil
	}
}

func (s *ShutdownService) Close(_ context.Context) error {
	close(s.done)
	return nil
}
