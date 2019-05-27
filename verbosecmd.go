package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

// VerboseCmd represents a system command with verbose output.
type VerboseCmd struct {
	Cmd exec.Cmd
}

// NewVerboseCmd creates a new instance of VerboseCmd.
func NewVerboseCmd(name string, args ...string) VerboseCmd {
	cmd := exec.Command(name, args...)
	return VerboseCmd{*cmd}
}

//Run executes the command and returns any errors in case exist.
func (v VerboseCmd) Run() error {
	var stderr bytes.Buffer
	v.Cmd.Stderr = &stderr
	err := v.Cmd.Run()
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr.String())
	}
	return nil
}
