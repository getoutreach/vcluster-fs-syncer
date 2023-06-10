// Copyright 2023 Outreach Corporation. All Rights Reserved.

// Description: Implements bind mount logic for Darwin. Currently,
// is a no-op.

package syncer

// bindMount bind mounts from onto to, thus files written from to go into from
// and vice-versa
func bindMount(from, to string) error {
	return nil
}

func unmountBind(dir string) error {
	return nil
}
