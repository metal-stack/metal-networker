{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.IfacesData*/ -}}
{{ .Comment }}
[Match]
Name=lan{{ .Index }}

[Network]
{{- range .EVPNIfaces }}
VXLAN=vni{{ .VXLAN.ID }}
{{- end }}