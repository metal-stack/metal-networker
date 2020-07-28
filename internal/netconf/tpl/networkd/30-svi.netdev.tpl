{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.EVPNIface*/ -}}
{{ .SVI.Comment }}
[NetDev]
Name=vlan{{ .VRF.ID }}
Kind=vlan

[VLAN]
Id={{ .SVI.VLANID }}
