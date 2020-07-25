{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.TenantData*/ -}}
{{ .SVI.Comment }}
[NetDev]
Name=vlan{{ .VRF.ID }}
Kind=vlan

[VLAN]
Id={{ .SVI.VLANID }}
