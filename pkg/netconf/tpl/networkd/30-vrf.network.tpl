{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.EVPNIface*/ -}}
{{ .VRF.Comment }}
[Match]
Name=vrf{{ .VRF.ID }}