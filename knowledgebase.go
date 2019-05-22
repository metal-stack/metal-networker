package main

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

const VlanOffset = 1000

// Generated with: https://mengzhuo.github.io/yaml-to-go/.
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

func NewKnowledgeBase(path string) KnowledgeBase {
	d := mustUnmarshal(path)
	d.fillVLANIDs()
	return d
}

func mustUnmarshal(path string) KnowledgeBase {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	d := &KnowledgeBase{}
	err = yaml.Unmarshal(f, &d)
	if err != nil {
		panic(err)
	}
	return *d
}

func (kb *KnowledgeBase) fillVLANIDs() {
	for i := 0; i < len(kb.Networks); i++ {
		kb.Networks[i].Vlan = VlanOffset + i
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
	result := false
	for _, n := range kb.Networks {
		if n.Primary {
			if result {
				result = false
				break
			}
			result = true
		}
	}
	return result
}

func (kb *KnowledgeBase) containsOneUnderlay() bool {
	result := false
	for _, n := range kb.Networks {
		if n.Underlay {
			if result {
				result = false
				break
			}
			result = true
		}
	}
	return result
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

func (kb *KnowledgeBase) mustGetPrimary() Network {
	for _, n := range kb.Networks {
		if n.Primary {
			return n
		}
	}
	panic("no primary network available")
}

func (kb *KnowledgeBase) mustGetUnderlay() Network {
	for _, n := range kb.Networks {
		if n.Underlay {
			return n
		}
	}
	panic("no underlay network available")
}