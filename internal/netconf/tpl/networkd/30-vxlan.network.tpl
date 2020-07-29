{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.EVPNIface*/ -}}
{{ .VXLAN.Comment }}
[Match]
Name=vni{{ .VXLAN.ID }}

[Link]
MTUBytes=9000

[Network]
Bridge=bridge

[BridgeVLAN]
PVID={{ .SVI.VLANID }}
EgressUntagged={{ .SVI.VLANID }}
