// Copyright 2023 Outreach Corporation. All Rights Reserved.

// Description: Implements bind mount logic for Linux.

package syncer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/moby/sys/mount"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Sadly golang/sys doesn't have UmountNoFollow although it's there since Linux 2.6.34
const UmountNoFollow = 0x8

// bindMount bind mounts from onto to, thus files written from to go into from
// and vice-versa
func bindMount(from, to string) error {
	absFrom, err := filepath.EvalSymlinks(from)
	if err != nil {
		return fmt.Errorf("Could not resolve symlink for from %v", from)
	}

	fromSt, err := os.Stat(from)
	if err != nil {
		return errors.Wrap(err, "failed to stat from")
	}

	if err := os.Mkdir(to, fromSt.Mode()); os.IsExist(err) {
		err = unmountBind(from)
		if err != nil {
			// TODO: better way to handle this?
			logrus.WithError(err).Warn("failed to unmount dir for remount")
		}
	} else if err != nil {
		return errors.Wrap(err, "failed to create to dir")
	}

	return errors.Wrap(
		mount.Mount(absFrom, to, "bind", "rbind,rw"),
		"failed to bind mount",
	)
}

func unmountBind(dir string) error {
	if err := mount.Unmount(dir); err != nil {
		return errors.Wrapf(err, "failed to unmount %s", dir)
	}

	return os.Remove(dir)
}
