// Copyright 2024 Outreach Corporation. All Rights Reserved.

// Description: This file exposes the private HTTP service for vcluster-fs-syncer.
// Managed: true

package vcluster_fs_syncer //nolint:revive // Why: We allow [-_].

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getoutreach/httpx/pkg/handlers"
	// Place any extra imports for your service code here
	// <<Stencil::Block(imports)>>
	// <</Stencil::Block>>
)

// PrivateHTTPDependencies is used to inject dependencies into the HTTPService service
// activity. Great examples of integrations to be placed into here would be a database
// connection or perhaps a redis client that the service activity needs to use.
type PrivateHTTPDependencies struct {
	// <<Stencil::Block(privateHTTPDependencies)>>

	// <</Stencil::Block>>
}

// HTTPService handles internal http requests, suchs as metrics, health
// and readiness checks. This is required for ALL services to have.
type HTTPService struct {
	handlers.Service

	cfg  *Config
	deps *PrivateHTTPDependencies
}

// NewHTTPService creates a new HTTPService service activity.
func NewHTTPService(cfg *Config, deps *PrivateHTTPDependencies) *HTTPService {
	return &HTTPService{
		cfg:  cfg,
		deps: deps,
	}
}

// Run is the entrypoint for the HTTPService serviceActivity.
func (s *HTTPService) Run(ctx context.Context) error {
	// create a http handler (handlers.Service does metrics, health etc)
	// <<Stencil::Block(privatehandler)>>
	s.App = http.NotFoundHandler()
	// <</Stencil::Block>>
	return s.Service.Run(ctx, fmt.Sprintf("%s:%d", s.cfg.ListenHost, s.cfg.HTTPPort))
}
