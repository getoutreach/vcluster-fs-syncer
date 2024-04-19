// Copyright 2023 Outreach Corporation. All Rights Reserved.

// Description: Contains the SyncerService struct and its methods.

package vcluster_fs_syncer //nolint:revive // Why: We allow [-_].

import (
	"context"

	"github.com/getoutreach/vcluster-fs-syncer/internal/syncer"
	"github.com/sirupsen/logrus"
)

// SyncerService implements the ServiceActivity framework for
// the vcluster-fs-syncer service. This service activity is the
// wrapper for all logic that is required to run the service.
type SyncerService struct {
	syncer *syncer.Syncer
}

// NewSyncerService create a new SyncerService that implements
// the core of this service's logic.
func NewSyncerService(cfg *Config) *SyncerService {
	return &SyncerService{syncer.NewSyncer(cfg.FromPath, cfg.ToPath, logrus.New())}
}

// Run starts the SyncerService.
func (s *SyncerService) Run(ctx context.Context) error {
	return s.syncer.Start(ctx)
}

// Close shuts down the SyncerService.
func (s *SyncerService) Close(_ context.Context) error {
	return s.syncer.Close()
}
