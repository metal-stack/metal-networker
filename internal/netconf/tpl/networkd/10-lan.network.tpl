{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.LanNetworkData*/ -}}
{{ .Comment }}
[Match]
Name=lan{{ .Index }}

[Network]
{{- range .Tenants }}
VXLAN={{ .VXLAN.ID }}
{{- end }}