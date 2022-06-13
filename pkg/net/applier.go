package net

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"io"
	"os"
	"text/template"
)

// Reloader triggers the reload to carry out the changes of an applier.
type Reloader interface {
	Reload() error
}

// Validator is an interface to apply common validation.
type Validator interface {
	Validate() error
}

// NetworkApplier holds the toolset for applying network configuration changes.
type NetworkApplier struct {
	data      interface{}
	validator Validator
	reloader  Reloader
}

// NewNetworkApplier creates a new NewNetworkApplier.
func NewNetworkApplier(data interface{}, validator Validator, reloader Reloader) *NetworkApplier {
	return &NetworkApplier{data: data, validator: validator, reloader: reloader}
}

// Apply applies the current configuration with the given template.
//
func (n *NetworkApplier) Apply(tpl template.Template, tmpFile, destFile string, reload bool) (bool, error) {
	f, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return false, err
	}

	defer func() {
		_ = f.Close()
	}()

	w := bufio.NewWriter(f)
	err = n.Render(w, tpl)

	if err != nil {
		return false, err
	}

	err = w.Flush()
	if err != nil {
		return false, err
	}

	err = n.validate()
	if err != nil {
		return false, err
	}

	equal := n.compare(tmpFile, destFile)
	if equal {
		return false, nil
	}

	err = os.Rename(tmpFile, destFile)
	if err != nil {
		return false, err
	}

	if !reload {
		return true, nil
	}

	err = n.reload()
	if err != nil {
		return true, err
	}

	return true, nil
}

// Render renders the network interfaces to the given writer using the given template.
func (n *NetworkApplier) Render(w io.Writer, tpl template.Template) error {
	return tpl.Execute(w, n.data)
}

func (n *NetworkApplier) validate() error {
	return n.validator.Validate()
}

// reload reloads the necessary services when the network interfaces configuration was changed.
func (n *NetworkApplier) reload() error {
	return n.reloader.Reload()
}

// compare compare source and target for hash equality.
func (n *NetworkApplier) compare(source, target string) bool {
	sourceChecksum, err := checksum(source)
	if err != nil {
		return false
	}

	targetChecksum, err := checksum(target)
	if err != nil {
		return false
	}

	return bytes.Equal(sourceChecksum, targetChecksum)
}

func checksum(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = f.Close()
	}()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
