{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.SystemdNetworkData*/ -}}
{{ .Comment }}
[Match]
Name=lan{{ .Index }}

[Network]
DHCP=ipv6