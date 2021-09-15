// Code managed by bootstrap.  Only edit code within the ///Block()/EndBlock() zones.

// Package main implements the main entrypoint for the vcluster-fs-syncer service.
//
// To build this package do:
//
//   make
//
// To run this do:
//
//   ./bin/vcluster-fs-syncer
//
// To run with honeycomb enabled do: (Note: the below section assumes you have go-outreach cloned somewhere.)
//
//    $> push <go-outreach>
//    $> ./scripts/devconfig.sh
//    $> vault kv get -format=json dev/honeycomb/dev-env | jq -cr '.data.data.apiKey' > ~/.outreach/honeycomb.key
//    $> popd
//    $> ./scripts/devconfig.sh
//    $> ./bin/vcluster-fs-syncer
//
package main

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/automaxprocs/maxprocs"
	"golang.org/x/sync/errgroup"

	"github.com/getoutreach/gobox/pkg/app"
	"github.com/getoutreach/gobox/pkg/env"
	"github.com/getoutreach/gobox/pkg/events"
	"github.com/getoutreach/gobox/pkg/log"
	"github.com/getoutreach/gobox/pkg/orerr"
	"github.com/getoutreach/gobox/pkg/trace"

	vcluster_fs_syncer "github.com/getoutreach/vcluster-fs-syncer/internal/vcluster-fs-syncer"
	// Place any extra imports for your startup code here
	///Block(imports)
	///EndBlock(imports)
)

func setMaxProcs(ctx context.Context) func() {
	// Set GOMAXPROCS to match the Linux container CPU quota (if any)
	undo, err := maxprocs.Set(maxprocs.Logger(func(m string, args ...interface{}) {
		message := fmt.Sprintf(m, args...)
		log.Info(ctx, "maxprocs.Set", log.F{"message": message})
	}))
	if err != nil {
		log.Error(ctx, "maxprocs.Set", events.NewErrorInfo(err))
		return func() {}
	}
	return undo
}

type serviceActivity interface {
	Run(ctx context.Context, config *vcluster_fs_syncer.Config) error
	Close(ctx context.Context) error
}

// Place any customized code for your service in this block
///Block(customized)
///EndBlock(customized)

func main() { //nolint: funlen
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	env.ApplyOverrides()
	app.SetName("vcluster-fs-syncer")
	defer setMaxProcs(ctx)()

	cfg := vcluster_fs_syncer.LoadConfig(ctx)

	if err := trace.InitTracer(ctx, "vcluster-fs-syncer"); err != nil {
		log.Error(ctx, "tracing failed to start", events.NewErrorInfo(err))
		return
	}
	defer trace.CloseTracer(ctx)

	log.Info(ctx, "starting", app.Info(), cfg, log.F{"app.pid": os.Getpid()})

	// Place any code for your service to run before registering service activities in this block
	///Block(initialization)
	///EndBlock(initialization)

	acts := []serviceActivity{
		vcluster_fs_syncer.NewShutdownService(),
		&vcluster_fs_syncer.HTTPService{},
		// Place any additional ServiceActivities that your service has built here to have them handled automatically
		///Block(services)
		&vcluster_fs_syncer.SyncerService{},
		///EndBlock(services)
	}

	// Place any code for your service to run during startup in this block
	///Block(startup)
	///EndBlock(startup)

	ctx, cancelWithError := orerr.CancelWithError(ctx)
	g, ctx := errgroup.WithContext(ctx)
	for idx := range acts {
		act := acts[idx]
		g.Go(func() error {
			defer act.Close(ctx)
			err := act.Run(ctx, cfg)
			if err != nil {
				cancelWithError(err)
			}
			return err
		})
	}
	if err := g.Wait(); err != nil {
		log.Info(ctx, "Closing down service due to", events.NewErrorInfo(err))
	}
}
