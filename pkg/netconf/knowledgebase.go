package netconf

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/metal-stack/metal-hammer/pkg/api"
	"go.uber.org/zap"

	"github.com/metal-stack/metal-go/api/models"
	mn "github.com/metal-stack/metal-lib/pkg/net"
	"github.com/metal-stack/v"

	"gopkg.in/yaml.v3"
)

const (
	// VLANOffset defines a number to start with when creating new VLAN IDs.
	VLANOffset = 1000
)

type (
	// config was generated with: https://mengzhuo.github.io/yaml-to-go/.
	// It represents the input yaml that is needed to render network configuration files.
	config struct {
		api.InstallerConfig
		log *zap.SugaredLogger
	}
)

// New creates a new instance of this type.
func New(log *zap.SugaredLogger, path string) (*config, error) {
	log.Infof("loading: %s", path)

	f, err := os.ReadFile(path)
	if err != nil {
		log.Panic(err)
	}

	installer := &api.InstallerConfig{}
	err = yaml.Unmarshal(f, &installer)

	if err != nil {
		return nil, err
	}

	return &config{
		InstallerConfig: *installer,
		log:             log,
	}, nil
}

// Validate validates the containing information depending on the demands of the bare metal type.
func (c config) Validate(kind BareMetalType) error {
	if len(c.Networks) == 0 {
		return errors.New("expectation at least one network is present failed")
	}

	if !c.containsSinglePrivatePrimary() {
		return errors.New("expectation exactly one 'private: true' network is present failed")
	}

	if kind == Firewall {
		if !c.allNonUnderlayNetworksHaveNonZeroVRF() {
			return errors.New("networks with 'underlay: false' must contain a value vor 'vrf' as it is used for BGP")
		}

		if !c.containsSingleUnderlay() {
			return errors.New("expectation exactly one underlay network is present failed")
		}

		if !c.containsAnyPublicNetwork() {
			return errors.New("expectation at least one public network (private: false, " +
				"underlay: false) is present failed")
		}

		for _, net := range c.GetNetworks(mn.External) {
			if len(net.Destinationprefixes) == 0 {
				return errors.New("non-private, non-underlay networks must contain destination prefix(es) to make " +
					"any sense of it")
			}
		}

		if c.isAnyNAT() && len(c.getPrivatePrimaryNetwork().Prefixes) == 0 {
			return errors.New("private network must not lack prefixes since nat is required")
		}
	}

	net := c.getPrivatePrimaryNetwork()

	if kind == Firewall {
		net = c.getUnderlayNetwork()
	}

	if len(net.Ips) == 0 {
		return errors.New("at least one IP must be present to be considered as LOOPBACK IP (" +
			"'private: true' network IP for machine, 'underlay: true' network IP for firewall")
	}

	if net.Asn != nil && *net.Asn <= 0 {
		return errors.New("'asn' of private (machine) resp. underlay (firewall) network must not be missing")
	}

	if len(c.Nics) == 0 {
		return errors.New("at least one 'nics/nic' definition must be present")
	}

	if !c.nicsContainValidMACs() {
		return errors.New("each 'nic' definition must contain a valid 'mac'")
	}

	return nil
}

func (c config) containsAnyPublicNetwork() bool {
	if len(c.GetNetworks(mn.External)) > 0 {
		return true
	}
	for _, n := range c.Networks {
		if isDMZNetwork(n) {
			return true
		}
	}
	return false
}

func (c config) containsSinglePrivatePrimary() bool {
	return c.containsSingleNetworkOf(mn.PrivatePrimaryUnshared) != c.containsSingleNetworkOf(mn.PrivatePrimaryShared)
}

func (c config) containsSingleUnderlay() bool {
	return c.containsSingleNetworkOf(mn.Underlay)
}

func (c config) containsSingleNetworkOf(t string) bool {
	possibleNetworks := c.GetNetworks(t)
	return len(possibleNetworks) == 1
}

// CollectIPs collects IPs of the given networks.
func (c config) CollectIPs(types ...string) []string {
	var result []string

	networks := c.GetNetworks(types...)
	for _, network := range networks {
		result = append(result, network.Ips...)
	}

	return result
}

// GetNetworks returns all networks present.
func (c config) GetNetworks(types ...string) []*models.V1MachineNetwork {
	var result []*models.V1MachineNetwork

	for _, t := range types {
		for _, n := range c.Networks {
			if n.Networktype == nil {
				continue
			}
			if *n.Networktype == t {
				result = append(result, n)
			}
		}
	}

	return result
}

func (c config) isAnyNAT() bool {
	for _, net := range c.Networks {
		if net.Nat != nil && *net.Nat {
			return true
		}
	}

	return false
}

func (c config) getPrivatePrimaryNetwork() *models.V1MachineNetwork {
	return c.GetNetworks(mn.PrivatePrimaryUnshared, mn.PrivatePrimaryShared)[0]
}

func (c config) getUnderlayNetwork() *models.V1MachineNetwork {
	// Safe access since validation ensures there is exactly one.
	return c.GetNetworks(mn.Underlay)[0]
}

func (c config) GetDefaultRouteNetwork() *models.V1MachineNetwork {
	externalNets := c.GetNetworks(mn.External)
	for _, network := range externalNets {
		if containsDefaultRoute(network.Destinationprefixes) {
			return network
		}
	}

	privateSecondarySharedNets := c.GetNetworks(mn.PrivateSecondaryShared)
	for _, network := range privateSecondarySharedNets {
		if containsDefaultRoute(network.Destinationprefixes) {
			return network
		}
	}

	return nil
}

func (c config) getDefaultRouteVRFName() (string, error) {
	if network := c.GetDefaultRouteNetwork(); network != nil {
		return vrfNameOf(network), nil
	}

	return "", fmt.Errorf("there is no network providing a default (0.0.0.0/0) route")
}

func (c config) nicsContainValidMACs() bool {
	for _, nic := range c.Nics {
		if nic.Mac == nil || *nic.Mac == "" {
			return false
		}

		if _, err := net.ParseMAC(*nic.Mac); err != nil {
			c.log.Errorf("invalid mac: %s", nic.Mac)
			return false
		}
	}

	return true
}

func (c config) allNonUnderlayNetworksHaveNonZeroVRF() bool {
	for _, net := range c.Networks {
		if net.Underlay != nil && *net.Underlay {
			continue
		}

		if net.Vrf != nil && *net.Vrf <= 0 {
			return false
		}
	}

	return true
}

func versionHeader(uuid string) string {
	version := v.V.String()
	if os.Getenv("GO_ENV") == "testing" {
		version = ""
	}
	return fmt.Sprintf("# This file was auto generated for machine: '%s' by app version %s.\n# Do not edit.",
		uuid, version)
}
