package netconf

import "net"

func getLocalBGPIP(primary Network) string {
	firstIP := net.ParseIP(primary.Ips[0])
	for _, prefix := range primary.Prefixes {
		ip, network, err := net.ParseCIDR(prefix)
		if err != nil {
			continue
		}
		if network.Contains(firstIP) {
			// Set the last octet to ".0" (sometimes referred to as "network identifier") to ensure a free IP within
			// this network. It is absolutely fine to use that IP. Even Amazon EC2 is assigning it to machines.
			ip[len(ip)-1] = 0
			return ip.String()
		}
	}
	return ""
}
