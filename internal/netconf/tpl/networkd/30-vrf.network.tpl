{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.TenantData*/ -}}
{{ .VRF.Comment }}
[Match]
Name=vrf{{ .VRF.ID }}