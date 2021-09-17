package vcluster_fs_syncer //nolint:revive

import (
	"context"

	"github.com/getoutreach/vcluster-fs-syncer/internal/syncer"
	"github.com/sirupsen/logrus"
)

type SyncerService struct {
	syncer *syncer.Syncer
}

func (s *SyncerService) Run(ctx context.Context, conf *Config) error {
	s.syncer = syncer.NewSyncer(conf.FromPath, conf.ToPath, logrus.New())
	return s.syncer.Start(ctx)
}

func (s *SyncerService) Close(_ context.Context) error {
	return s.syncer.Close()
}
