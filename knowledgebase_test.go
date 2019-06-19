package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func mustNewKnowledgeBase(t *testing.T) KnowledgeBase {
	assert := assert.New(t)

	d := NewKnowledgeBase("testdata/install.yaml")
	assert.NotNil(d)

	return d
}

func TestNewKnowledgeBase(t *testing.T) {
	assert := assert.New(t)

	d := mustNewKnowledgeBase(t)

	assert.Equal("firewall", d.Hostname)
	assert.NotEmpty(d.Networks)
	assert.Equal(3, len(d.Networks))

	// primary network
	n := d.Networks[0]
	assert.Equal(1, len(n.Ips))
	assert.Equal("10.0.16.2", n.Ips[0])
	assert.Equal(1, len(n.Prefixes))
	assert.Equal("10.0.16.0/22", n.Prefixes[0])
	assert.True(n.Primary)
	assert.Equal(3981, n.Vrf)

	// external network
	n = d.Networks[1]
	assert.Equal(1, len(n.Destinationprefixes))
	assert.Equal("0.0.0.0/0", n.Destinationprefixes[0])
	assert.Equal(1, len(n.Ips))
	assert.Equal("185.1.2.3", n.Ips[0])
	assert.Equal(2, len(n.Prefixes))
	assert.Equal("185.1.2.0/24", n.Prefixes[0])
	assert.Equal("185.27.0.0/22", n.Prefixes[1])
	assert.False(n.Underlay)
	assert.False(n.Primary)
	assert.True(n.Nat)
	assert.Equal(104009, n.Vrf)

	// underlay network
	n = d.Networks[2]
	assert.Equal(int64(4200003073), n.Asn)
	assert.Equal(1, len(n.Ips))
	assert.Equal("10.1.0.1", n.Ips[0])
	assert.Equal(1, len(n.Prefixes))
	assert.Equal("10.0.12.0/22", n.Prefixes[0])
	assert.True(n.Underlay)
}

func TestValidationSucceeds(t *testing.T) {
	assert := assert.New(t)

	d := mustNewKnowledgeBase(t)
	v := d.validate()
	assert.NoError(v)
}

func TestValidationFailsForUnderlay(t *testing.T) {
	assert := assert.New(t)
	d := mustNewKnowledgeBase(t)

	for i := 0; i < len(d.Networks); i++ {
		d.Networks[i].Underlay = false
	}
	err := d.validate()
	assert.Error(err, "missing validation error for absent underlay")

	for i := 0; i < len(d.Networks); i++ {
		d.Networks[i].Underlay = true
	}
	err = d.validate()
	assert.Error(err, "missing validation error for multiple underlays failed")
}

func TestValidationFailsForPrimary(t *testing.T) {
	assert := assert.New(t)
	d := mustNewKnowledgeBase(t)

	for i := 0; i < len(d.Networks); i++ {
		d.Networks[i].Primary = false
	}
	err := d.validate()
	assert.Error(err, "missing validation error for absent primary")

	for i := 0; i < len(d.Networks); i++ {
		d.Networks[i].Primary = true
	}
	err = d.validate()
	assert.Error(err, "missing validation error for multiple primaries failed")
}

func TestValidationFailsForMissingLoopbackIp(t *testing.T) {
	assert := assert.New(t)
	d := mustNewKnowledgeBase(t)

	for i, n := range d.Networks {
		if n.Underlay {
			d.Networks[i].Ips = []string{}
		}
	}
	err := d.validate()
	assert.Error(err, "missing validation error for absent loopback ip")
}
