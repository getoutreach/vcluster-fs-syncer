package syncer

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

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
		err = syscall.Unmount(absFrom, syscall.MNT_DETACH|UmountNoFollow)
		if err != nil {
			// TODO: better way to handle this?
			logrus.WithError(err).Warn("failed to unmount dir for remount")
		}
	} else if err != nil {
		return errors.Wrap(err, "failed to create to dir")
	}

	if err := syscall.Mount(absFrom, to, "bind", syscall.MS_BIND, ""); err != nil {
		return fmt.Errorf("Could not bind mount %v to %v: %v", absFrom, to, err)
	}

	if err := syscall.Mount("none", to, "", syscall.MS_SHARED, ""); err != nil {
		return fmt.Errorf("Could not make mount point %v %s: %v", to, syscall.MS_SHARED, err)
	}

	return nil
}
