package net

import (
	"context"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/coreos/go-systemd/v22/unit"
)

// Validator is an interface to apply common validation.
type Validator interface {
	Validate() error
}

// DBusTemplateValidator starts a dbus template (templatename@instancename.service) to apply validation.
type DBusTemplateValidator struct {
	TemplateName, InstanceName string
}

// Validate applies validation by starting a dbus templated instance.
func (v DBusTemplateValidator) Validate() error {
	dbc, err := dbus.NewWithContext(context.Background())
	if err != nil {
		return fmt.Errorf("unable to connect to dbus: %w", err)
	}
	defer dbc.Close()

	c := make(chan string)
	u := fmt.Sprintf("%s@%s.service", v.TemplateName, unit.UnitNamePathEscape(v.InstanceName))
	_, err = dbc.StartUnitContext(context.Background(), u, "replace", c)

	if err != nil {
		return err
	}

	job := <-c
	if job != done {
		return fmt.Errorf("validation failed %s", job)
	}

	return nil
}
