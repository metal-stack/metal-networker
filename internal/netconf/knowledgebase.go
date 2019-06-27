package netconf

import (
	"errors"
	"fmt"
	"io/ioutil"

	"git.f-i-ts.de/cloud-native/metallib/zapup"

	"github.com/metal-pod/v"

	"gopkg.in/yaml.v3"
)

var log = zapup.MustRootLogger().Sugar()

// VLANOffset defines a number to start with when creating new VLAN IDs.
const VLANOffset = 1000

// NetworkType represents the functional type of a network.
type NetworkType int

const (
	// Underlay represents the Underlay network.
	Underlay NetworkType = iota
	// Primary represents the primary network.
	Primary
	// External represents the external network.
	External
)

// KnowledgeBase was generated with: https://mengzhuo.github.io/yaml-to-go/.
// It represents the input yaml that is needed to render network configuration files.
type KnowledgeBase struct {
	Hostname     string    `yaml:"hostname"`
	Ipaddress    string    `yaml:"ipaddress"`
	Asn          string    `yaml:"asn"`
	Networks     []Network `yaml:"networks"`
	Machineuuid  string    `yaml:"machineuuid"`
	Sshpublickey string    `yaml:"sshpublickey"`
	Password     string    `yaml:"password"`
	Devmode      bool      `yaml:"devmode"`
	Console      string    `yaml:"console"`
}

// Network is a representation of a tenant network.
type Network struct {
	Asn                 int64    `yaml:"asn"`
	Destinationprefixes []string `yaml:"destinationprefixes"`
	Ips                 []string `yaml:"ips"`
	Nat                 bool     `yaml:"nat"`
	Networkid           string   `yaml:"networkid"`
	Prefixes            []string `yaml:"prefixes"`
	Primary             bool     `yaml:"primary"`
	Underlay            bool     `yaml:"underlay"`
	Vrf                 int      `yaml:"vrf"`
	Vlan                int      `yaml:"vlan,omitempty"`
}

// NewKnowledgeBase creates a new instance of this type.
func NewKnowledgeBase(path string) KnowledgeBase {
	d := mustUnmarshal(path)
	d.fillVLANIDs()
	return d
}

func mustUnmarshal(path string) KnowledgeBase {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		log.Panic(err)
	}

	d := &KnowledgeBase{}
	err = yaml.Unmarshal(f, &d)
	if err != nil {
		log.Panic(err)
	}
	return *d
}

func (kb *KnowledgeBase) fillVLANIDs() {
	for i := 0; i < len(kb.Networks); i++ {
		kb.Networks[i].Vlan = VLANOffset + i
	}
}

func (kb *KnowledgeBase) validate() error {
	b := kb.containsAsn()
	if !b {
		return errors.New("asn must not be missing")
	}
	b = kb.containsAtLeastOneLoopbackIP()
	if !b {
		return errors.New("underlay ip(s) must not be absent")
	}
	b = kb.containsOnePrimary()
	if !b {
		return errors.New("expectation exactly one primary network is present failed")
	}
	b = kb.containsOneUnderlay()
	if !b {
		return errors.New("expectation exactly one underlay network is present failed")
	}
	return nil
}

func (kb *KnowledgeBase) containsOnePrimary() bool {
	return kb.containsSingleNetworkOf(Primary)
}

func (kb *KnowledgeBase) containsOneUnderlay() bool {
	return kb.containsSingleNetworkOf(Underlay)
}

func (kb *KnowledgeBase) containsSingleNetworkOf(networkType NetworkType) bool {
	possibleNetworks := kb.GetNetworks(networkType)
	return len(possibleNetworks) == 1
}

func (kb *KnowledgeBase) containsAtLeastOneLoopbackIP() bool {
	for _, n := range kb.Networks {
		if n.Underlay && len(n.Ips) >= 1 {
			return true
		}
	}
	return false
}

func (kb *KnowledgeBase) containsAsn() bool {
	for _, n := range kb.Networks {
		if n.Underlay && n.Asn > 0 {
			return true
		}
	}
	return false
}

func (kb *KnowledgeBase) GetNetworks(networkType ...NetworkType) []Network {
	var result []Network
	for _, t := range networkType {
		for _, n := range kb.Networks {
			switch t {
			case Primary:
				if n.Primary {
					result = append(result, n)
				}
			case Underlay:
				if n.Underlay {
					result = append(result, n)
				}
			case External:
				if !n.Underlay && !n.Primary {
					result = append(result, n)
				}
			}
		}
	}
	return result
}

func (kb KnowledgeBase) getPrimaryNetwork() Network {
	// Safe access since a priory validation ensures there is exactly one.
	return kb.GetNetworks(Primary)[0]
}

func (kb KnowledgeBase) getUnderlayNetwork() Network {
	// Safe access since a priory validation ensures there is exactly one.
	return kb.GetNetworks(Underlay)[0]
}

func versionHeader(uuid string) string {
	return fmt.Sprintf("# This file was auto generated for machine: '%s' by app version %s.\n# Do not edit.",
		uuid, v.V.String())
}
