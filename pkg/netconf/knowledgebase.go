package netconf

import (
	"errors"
	"fmt"
	"net"
	"os"

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
	// KnowledgeBase was generated with: https://mengzhuo.github.io/yaml-to-go/.
	// It represents the input yaml that is needed to render network configuration files.
	KnowledgeBase struct {
		Hostname     string                    `yaml:"hostname"`
		Ipaddress    string                    `yaml:"ipaddress"`
		Asn          string                    `yaml:"asn"`
		Networks     []models.V1MachineNetwork `yaml:"networks"`
		Machineuuid  string                    `yaml:"machineuuid"`
		Sshpublickey string                    `yaml:"sshpublickey"`
		Password     string                    `yaml:"password"`
		Devmode      bool                      `yaml:"devmode"`
		Console      string                    `yaml:"console"`
		Nics         []NIC                     `yaml:"nics"`
		VPN          *models.V1MachineVPN      `yaml:"vpn"`
	}

	// NIC is a representation of network interfaces attributes.
	NIC struct {
		Mac       string `yaml:"mac"`
		Name      string `yaml:"name"`
		Neighbors []struct {
			Mac       string        `yaml:"mac"`
			Name      interface{}   `yaml:"name"`
			Neighbors []interface{} `yaml:"neighbors"`
		} `yaml:"neighbors"`
	}
)

// NewKnowledgeBase creates a new instance of this type.
func NewKnowledgeBase(path string) KnowledgeBase {
	log.Infof("loading: %s", path)

	f, err := os.ReadFile(path)
	if err != nil {
		log.Panic(err)
	}

	kb := &KnowledgeBase{}
	err = yaml.Unmarshal(f, &kb)

	if err != nil {
		log.Panic(err)
	}

	return *kb
}

// Validate validates the containing information depending on the demands of the bare metal type.
func (kb KnowledgeBase) Validate(kind BareMetalType) error {
	if len(kb.Networks) == 0 {
		return errors.New("expectation at least one network is present failed")
	}

	if !kb.containsSinglePrivatePrimary() {
		return errors.New("expectation exactly one 'private: true' network is present failed")
	}

	if kind == Firewall {
		if !kb.allNonUnderlayNetworksHaveNonZeroVRF() {
			return errors.New("networks with 'underlay: false' must contain a value vor 'vrf' as it is used for BGP")
		}

		if !kb.containsSingleUnderlay() {
			return errors.New("expectation exactly one underlay network is present failed")
		}

		if !kb.containsAnyPublicNetwork() {
			return errors.New("expectation at least one public network (private: false, " +
				"underlay: false) is present failed")
		}

		for _, net := range kb.GetNetworks(mn.External) {
			if len(net.Destinationprefixes) == 0 {
				return errors.New("non-private, non-underlay networks must contain destination prefix(es) to make " +
					"any sense of it")
			}
		}

		if kb.isAnyNAT() && len(kb.getPrivatePrimaryNetwork().Prefixes) == 0 {
			return errors.New("private network must not lack prefixes since nat is required")
		}
	}

	net := kb.getPrivatePrimaryNetwork()

	if kind == Firewall {
		net = kb.getUnderlayNetwork()
	}

	if len(net.Ips) == 0 {
		return errors.New("at least one IP must be present to be considered as LOOPBACK IP (" +
			"'private: true' network IP for machine, 'underlay: true' network IP for firewall")
	}

	if net.Asn != nil && *net.Asn <= 0 {
		return errors.New("'asn' of private (machine) resp. underlay (firewall) network must not be missing")
	}

	if len(kb.Nics) == 0 {
		return errors.New("at least one 'nics/nic' definition must be present")
	}

	if !kb.nicsContainValidMACs() {
		return errors.New("each 'nic' definition must contain a valid 'mac'")
	}

	return nil
}

func (kb KnowledgeBase) containsAnyPublicNetwork() bool {
	if len(kb.GetNetworks(mn.External)) > 0 {
		return true
	}
	for _, n := range kb.Networks {
		if isDMZNetwork(n) {
			return true
		}
	}
	return false
}

func (kb KnowledgeBase) containsSinglePrivatePrimary() bool {
	return kb.containsSingleNetworkOf(mn.PrivatePrimaryUnshared) != kb.containsSingleNetworkOf(mn.PrivatePrimaryShared)
}

func (kb KnowledgeBase) containsSingleUnderlay() bool {
	return kb.containsSingleNetworkOf(mn.Underlay)
}

func (kb KnowledgeBase) containsSingleNetworkOf(t string) bool {
	possibleNetworks := kb.GetNetworks(t)
	return len(possibleNetworks) == 1
}

// CollectIPs collects IPs of the given networks.
func (kb KnowledgeBase) CollectIPs(types ...string) []string {
	var result []string

	networks := kb.GetNetworks(types...)
	for _, network := range networks {
		result = append(result, network.Ips...)
	}

	return result
}

// GetNetworks returns all networks present.
func (kb KnowledgeBase) GetNetworks(types ...string) []models.V1MachineNetwork {
	var result []models.V1MachineNetwork

	for _, t := range types {
		for _, n := range kb.Networks {
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

func (kb KnowledgeBase) isAnyNAT() bool {
	for _, net := range kb.Networks {
		if net.Nat != nil && *net.Nat {
			return true
		}
	}

	return false
}

func (kb KnowledgeBase) getPrivatePrimaryNetwork() models.V1MachineNetwork {
	return kb.GetNetworks(mn.PrivatePrimaryUnshared, mn.PrivatePrimaryShared)[0]
}

func (kb KnowledgeBase) getUnderlayNetwork() models.V1MachineNetwork {
	// Safe access since validation ensures there is exactly one.
	return kb.GetNetworks(mn.Underlay)[0]
}

func (kb KnowledgeBase) GetDefaultRouteNetwork() *models.V1MachineNetwork {
	externalNets := kb.GetNetworks(mn.External)
	for _, network := range externalNets {
		if containsDefaultRoute(network.Destinationprefixes) {
			return &network
		}
	}

	privateSecondarySharedNets := kb.GetNetworks(mn.PrivateSecondaryShared)
	for _, network := range privateSecondarySharedNets {
		if containsDefaultRoute(network.Destinationprefixes) {
			return &network
		}
	}

	return nil
}

func (kb KnowledgeBase) getDefaultRouteVRFName() (string, error) {
	if network := kb.GetDefaultRouteNetwork(); network != nil {
		return vrfNameOf(*network), nil
	}

	return "", fmt.Errorf("there is no network providing a default (0.0.0.0/0) route")
}

func (kb KnowledgeBase) nicsContainValidMACs() bool {
	for _, nic := range kb.Nics {
		if nic.Mac == "" {
			return false
		}

		if _, err := net.ParseMAC(nic.Mac); err != nil {
			log.Errorf("invalid mac: %s", nic.Mac)
			return false
		}
	}

	return true
}

func (kb KnowledgeBase) allNonUnderlayNetworksHaveNonZeroVRF() bool {
	for _, net := range kb.Networks {
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
