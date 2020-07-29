{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.EVPNIface*/ -}}
{{ .VRF.Comment }}
[NetDev]
Name=vrf{{ .VRF.ID }}
Kind=vrf

[VRF]
Table={{ .VRF.Table }}