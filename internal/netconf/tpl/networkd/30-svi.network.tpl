{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.TenantData*/ -}}
{{ .SVI.Comment }}
[Match]
Name=vlan{{ .VRF.ID }}

[Link]
MTUBytes=9000

[Network]
VRF=vrf{{ .VRF.ID }}
{{- range .SVI.Addresses }}
Address={{ . }}
{{- end }}
