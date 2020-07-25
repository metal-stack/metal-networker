{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.IfacesData*/ -}}
{{ .Comment }}
[Match]
Name=bridge

[Network]
{{- range .Tenants }}
VLAN=vlan{{ .VRF.ID }}
{{- end }}

{{- range .Tenants }}
[BridgeVLAN]
VLAN={{ .SVI.VLANID }}

{{- end }}