package main

import (
	"errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

const VlanOffset = 1000

// Generated with: https://mengzhuo.github.io/yaml-to-go/.
type KnowledgeBase struct {
	Hostname  string `yaml:"hostname"`
	Ipaddress string `yaml:"ipaddress"`
	Asn       string `yaml:"asn"`
	Networks  []struct {
		Asn                 int64         `yaml:"asn"`
		Destinationprefixes []interface{} `yaml:"destinationprefixes"`
		Ips                 []string      `yaml:"ips"`
		Nat                 bool          `yaml:"nat"`
		Networkid           string        `yaml:"networkid"`
		Prefixes            []string      `yaml:"prefixes"`
		Primary             bool          `yaml:"primary"`
		Underlay            bool          `yaml:"underlay"`
		Vrf                 int           `yaml:"vrf"`
		Vlan                int           `yaml:"vlan,omitempty"`
	} `yaml:"networks"`
	Machineuuid  string `yaml:"machineuuid"`
	Sshpublickey string `yaml:"sshpublickey"`
	Password     string `yaml:"password"`
	Devmode      bool   `yaml:"devmode"`
	Console      string `yaml:"console"`
}

func NewKnowledgeBase(path string) (*KnowledgeBase, error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	d := &KnowledgeBase{}
	yaml.Unmarshal(f, &d)
	if err != nil {
		return nil, err
	}

	d.fillVLANIDs()
	return d, err
}

func (kb *KnowledgeBase) fillVLANIDs() {
	for i := 0; i < len(kb.Networks); i++ {
		kb.Networks[i].Vlan = VlanOffset + i
	}
}

func (kb *KnowledgeBase) validate() error {
	b := kb.containsAsn()
	if b == false {
		return errors.New("asn must not be missing")
	}
	b = kb.containsAtLeastOneLoopbackIp()
	if b == false {
		return errors.New("underlay ip(s) must not be absent")
	}
	b = kb.containsOnePrimary()
	if b == false {
		return errors.New("expectation exactly one primary network is present failed")
	}
	b = kb.containsOneUnderlay()
	if b == false {
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

func (kb *KnowledgeBase) containsAtLeastOneLoopbackIp() bool {
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
