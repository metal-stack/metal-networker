package netconf

import "net"

func getLocalBGPIP(primary Network) string {
	firstIP := net.ParseIP(primary.Ips[0])
	for _, prefix := range primary.Prefixes {
		ip, network, err := net.ParseCIDR(prefix)
		if err == nil && network.Contains(firstIP) {
			// Set the last octet to "0" regardless of version
			ip[len(ip)-1] = 0
			return ip.String()
		}
	}
	return ""
}
