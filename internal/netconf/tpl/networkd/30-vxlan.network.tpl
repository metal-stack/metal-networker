{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.TenantData*/ -}}
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
