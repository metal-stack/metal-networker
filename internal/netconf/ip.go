package netconf

import (
	"fmt"
	"net"
)

func getLocalBGPIP(primary Network) (string, error) {
	firstIP := net.ParseIP(primary.Ips[0])
	for _, prefix := range primary.Prefixes {
		ip, network, err := net.ParseCIDR(prefix)
		if err != nil {
			continue
		}
		if network.Contains(firstIP) {
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("failure to figure out a local BGP IP from primary network: %v", primary)
}
