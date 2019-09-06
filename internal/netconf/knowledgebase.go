package netconf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"

	"git.f-i-ts.de/cloud-native/metallib/zapup"

	"github.com/metal-pod/v"

	"gopkg.in/yaml.v3"
)

var log = zapup.MustRootLogger().Sugar()

const (
	// VLANOffset defines a number to start with when creating new VLAN IDs.
	VLANOffset = 1000
	// Underlay represents the fabric network where infrastructure switches and routers are placed in.
	Underlay NetworkType = iota
	// Private represents the local machine network where all machines of a project are placed in.
	Private
	// Public represents an external network a machine has access to.
	Public
)

type (
	// NetworkType represents the functional type of a network.
	NetworkType int

	// KnowledgeBase was generated with: https://mengzhuo.github.io/yaml-to-go/.
	// It represents the input yaml that is needed to render network configuration files.
	KnowledgeBase struct {
		Hostname     string    `yaml:"hostname"`
		Ipaddress    string    `yaml:"ipaddress"`
		Asn          string    `yaml:"asn"`
		Networks     []Network `yaml:"networks"`
		Machineuuid  string    `yaml:"machineuuid"`
		Sshpublickey string    `yaml:"sshpublickey"`
		Password     string    `yaml:"password"`
		Devmode      bool      `yaml:"devmode"`
		Console      string    `yaml:"console"`
		Nics         []NIC     `yaml:"nics"`
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

	// Network is a representation of a tenant network.
	Network struct {
		Asn                 int64    `yaml:"asn"`
		Destinationprefixes []string `yaml:"destinationprefixes"`
		Ips                 []string `yaml:"ips"`
		Nat                 bool     `yaml:"nat"`
		Networkid           string   `yaml:"networkid"`
		Prefixes            []string `yaml:"prefixes"`
		Private             bool     `yaml:"private"`
		Underlay            bool     `yaml:"underlay"`
		Vrf                 int      `yaml:"vrf"`
		Vlan                int      `yaml:"vlan,omitempty"`
	}
)

// NewKnowledgeBase creates a new instance of this type.
func NewKnowledgeBase(path string) KnowledgeBase {
	log.Infof("loading: %s", path)
	f, err := ioutil.ReadFile(path)
	if err != nil {
		log.Panic(err)
	}

	kb := &KnowledgeBase{}
	err = yaml.Unmarshal(f, &kb)
	if err != nil {
		log.Panic(err)
	}
	for i := 0; i < len(kb.Networks); i++ {
		kb.Networks[i].Vlan = VLANOffset + i
	}

	return *kb
}

// Validate validates the containing information depending on the demands of the bare metal type.
func (kb KnowledgeBase) Validate(kind BareMetalType) error {
	if len(kb.Networks) == 0 {
		return errors.New("expectation at least one network is present failed")
	}
	if !kb.containsSinglePrivate() {
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
		for _, net := range kb.GetNetworks(Public) {
			if len(net.Destinationprefixes) == 0 {
				return errors.New("non-private, non-underlay networks must contain destination prefix(es) to make " +
					"any sense of it")
			}
		}
		if kb.isAnyNAT() && len(kb.getPrivateNetwork().Prefixes) == 0 {
			return errors.New("private network must not lack prefixes since nat is required")
		}
	}
	net := kb.getPrivateNetwork()
	if kind == Firewall {
		net = kb.getUnderlayNetwork()
	}
	if len(net.Ips) == 0 {
		return errors.New("at least one IP must be present to be considered as LOOPBACK IP (" +
			"'private: true' network IP for machine, 'underlay: true' network IP for firewall")
	}
	if net.Asn <= 0 {
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
	return len(kb.GetNetworks(Public)) > 0
}

func (kb KnowledgeBase) containsSinglePrivate() bool {
	return kb.containsSingleNetworkOf(Private)
}

func (kb KnowledgeBase) containsSingleUnderlay() bool {
	return kb.containsSingleNetworkOf(Underlay)
}

func (kb KnowledgeBase) containsSingleNetworkOf(types NetworkType) bool {
	possibleNetworks := kb.GetNetworks(types)
	return len(possibleNetworks) == 1
}

// CollectIPs collects IPs of the given networks.
func (kb KnowledgeBase) CollectIPs(types ...NetworkType) []string {
	var result []string
	networks := kb.GetNetworks(types...)
	for _, network := range networks {
		result = append(result, network.Ips...)
	}
	return result
}

// GetNetworks returns all networks present.
func (kb KnowledgeBase) GetNetworks(types ...NetworkType) []Network {
	var result []Network
	for _, t := range types {
		for _, n := range kb.Networks {
			switch t {
			case Private:
				if n.Private {
					result = append(result, n)
				}
			case Underlay:
				if n.Underlay {
					result = append(result, n)
				}
			case Public:
				if !n.Underlay && !n.Private {
					result = append(result, n)
				}
			}
		}
	}
	return result
}

func (kb KnowledgeBase) isAnyNAT() bool {
	for _, net := range kb.Networks {
		if net.Nat {
			return true
		}
	}
	return false
}

func (kb KnowledgeBase) getPrivateNetwork() Network {
	// Safe access since a priory validation ensures there is exactly one.
	return kb.GetNetworks(Private)[0]
}

func (kb KnowledgeBase) getUnderlayNetwork() Network {
	// Safe access since validation ensures there is exactly one.
	return kb.GetNetworks(Underlay)[0]
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
		if net.Underlay {
			continue
		}
		if net.Vrf <= 0 {
			return false
		}
	}
	return true
}

func versionHeader(uuid string) string {
	return fmt.Sprintf("# This file was auto generated for machine: '%s' by app version %s.\n# Do not edit.",
		uuid, v.V.String())
}
