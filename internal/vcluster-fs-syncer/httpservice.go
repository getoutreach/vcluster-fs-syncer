// Code managed by Bootstrap.
//
// Please edit this to more accurately match the server implementation.

package vcluster_fs_syncer //nolint:revive

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getoutreach/httpx/pkg/handlers"
	// Place any extra imports for your service code here
	///Block(imports)
	///EndBlock(imports)
)

// HTTPService handles internal http requests
type HTTPService struct {
	handlers.Service
}

func (s *HTTPService) Run(ctx context.Context, config *Config) error {
	// create a http handler (handlers.Service does metrics, health etc)
	///Block(privatehandler)
	s.App = http.NotFoundHandler()
	///EndBlock(privatehandler)
	return s.Service.Run(ctx, fmt.Sprintf("%s:%d", config.ListenHost, config.HTTPPort))
}
