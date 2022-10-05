package net

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
)

const done = "done"

// Reloader triggers the reload to carry out the changes of an applier.
type Reloader interface {
	Reload() error
}

// NewDBusReloader is a reloader for systemd units with dbus.
func NewDBusReloader(service string) DBusReloader {
	return DBusReloader{service}
}

// DBusReloader applies a systemd unit reload to apply reloading.
type DBusReloader struct {
	ServiceFilename string
}

// Reload reloads a systemd unit.
func (r DBusReloader) Reload() error {
	ctx := context.Background()
	dbc, err := dbus.NewWithContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to connect to dbus: %w", err)
	}
	defer dbc.Close()

	c := make(chan string)
	_, err = dbc.ReloadUnitContext(ctx, r.ServiceFilename, "replace", c)

	if err != nil {
		return err
	}

	job := <-c
	if job != done {
		return fmt.Errorf("reloading failed %s", job)
	}

	return nil
}
