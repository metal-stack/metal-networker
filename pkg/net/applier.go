package net

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"io"
	"os"
	"text/template"
)

// Applier is an interface to render changes and reload services to apply them.
type Applier interface {
	Apply(tpl template.Template, tmpFile, destFile string, reload bool) (bool, error)
	Render(writer io.Writer, tpl template.Template) error
	Reload() error
	Validate() error
	Compare(tmpFile, destFile string) bool
}

// networkApplier holds the toolset for applying network configuration changes.
type networkApplier struct {
	data      any
	validator Validator
	reloader  Reloader
}

// NewNetworkApplier creates a new NewNetworkApplier.
func NewNetworkApplier(data any, validator Validator, reloader Reloader) Applier {
	return &networkApplier{data: data, validator: validator, reloader: reloader}
}

// Apply applies the current configuration with the given template.
func (n *networkApplier) Apply(tpl template.Template, tmpFile, destFile string, reload bool) (bool, error) {
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

	err = n.Validate()
	if err != nil {
		return false, err
	}

	equal := n.Compare(tmpFile, destFile)
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

	err = n.Reload()
	if err != nil {
		return true, err
	}

	return true, nil
}

// Render renders the network interfaces to the given writer using the given template.
func (n *networkApplier) Render(w io.Writer, tpl template.Template) error {
	return tpl.Execute(w, n.data)
}

// Validate applies the given validator to validate current changes.
func (n *networkApplier) Validate() error {
	return n.validator.Validate()
}

// Reload reloads the necessary services when the network interfaces configuration was changed.
func (n *networkApplier) Reload() error {
	return n.reloader.Reload()
}

// Compare compare source and target for hash equality.
func (n *networkApplier) Compare(source, target string) bool {
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
