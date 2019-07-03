{{- /*gotype: git.f-i-ts.de/cloud-native/metal/metal-networker/internal/netconf.SystemdNetworkData*/ -}}
{{ .Comment }}
[Match]
Name=lan{{ .Index }}

[Network]
DHCP=ipv6