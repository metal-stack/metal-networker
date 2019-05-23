package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

type VerboseCmd struct {
	Cmd exec.Cmd
}

func NewVerboseCmd(name string, args ...string) VerboseCmd {
	cmd := exec.Command(name, args...)
	return VerboseCmd{*cmd}
}

func (v VerboseCmd) Run() error {
	var stderr bytes.Buffer
	v.Cmd.Stderr = &stderr
	err := v.Cmd.Run()
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr.String())
	}
	return nil
}
