// Copyright 2025 Outreach Corporation. Licensed under the Apache License 2.0.

// Description: This file is the entrypoint for vcluster-fs-syncer.
// Managed: true

// Package main implements the main entrypoint for the vcluster-fs-syncer service.
package main

import (
	"context"
	"os"

	"github.com/getoutreach/gobox/pkg/app"
	"github.com/getoutreach/gobox/pkg/async"
	"github.com/getoutreach/gobox/pkg/env"
	"github.com/getoutreach/gobox/pkg/events"
	"github.com/getoutreach/gobox/pkg/log"
	"github.com/getoutreach/gobox/pkg/trace"
	"github.com/getoutreach/stencil-golang/pkg/serviceactivities/automemlimit"
	"github.com/getoutreach/stencil-golang/pkg/serviceactivities/gomaxprocs"
	"github.com/getoutreach/stencil-golang/pkg/serviceactivities/shutdown"

	// Place any extra imports for your startup code here
	// <<Stencil::Block(imports)>>
	vcluster_fs_syncer "github.com/getoutreach/vcluster-fs-syncer/internal/vcluster-fs-syncer"
	// <</Stencil::Block>>
)

// Place any customized code for your service in this block
//
// <<Stencil::Block(customized)>>

// <</Stencil::Block>>

// dependencies is a conglomerate struct used for injecting dependencies into service
// activities.
type dependencies struct {
	privateHTTP vcluster_fs_syncer.PrivateHTTPDependencies
	// <<Stencil::Block(customServiceActivityDependencyInjection)>>

	// <</Stencil::Block>>
}

// main is the entrypoint for the vcluster-fs-syncer service.
func main() { //nolint: funlen // Why: We can't dwindle this down anymore without adding complexity.
	exitCode := 1
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
		os.Exit(exitCode)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	env.ApplyOverrides()
	app.SetName("vcluster-fs-syncer")

	cfg, err := vcluster_fs_syncer.LoadConfig(ctx)
	if err != nil {
		log.Error(ctx, "failed to load config", events.NewErrorInfo(err))
		return
	}

	if err := trace.InitTracer(ctx, "vcluster-fs-syncer"); err != nil {
		log.Error(ctx, "tracing failed to start", events.NewErrorInfo(err))
		return
	}
	defer trace.CloseTracer(ctx)

	// Initialize variable for service activity dependency injection.
	var deps dependencies

	log.Info(ctx, "starting", app.Info(), cfg, log.F{"app.pid": os.Getpid()})

	// Place any code for your service to run before registering service activities in this block
	// <<Stencil::Block(initialization)>>
	sync := vcluster_fs_syncer.NewSyncerService(cfg)
	// <</Stencil::Block>>

	acts := []async.Runner{
		shutdown.New(),
		gomaxprocs.New(),
		automemlimit.New(),
		vcluster_fs_syncer.NewHTTPService(cfg, &deps.privateHTTP),

		// Place any additional ServiceActivities that your service has built here to have them handled automatically
		//
		// <<Stencil::Block(services)>>
		sync,
		// <</Stencil::Block>>
	}

	// Place any code for your service to run during startup in this block
	//
	// <<Stencil::Block(startup)>>

	// <</Stencil::Block>>

	err = async.RunGroup(acts).Run(ctx)
	if shutdown.HandleShutdownConditions(ctx, err) {
		exitCode = 0
	}
}
