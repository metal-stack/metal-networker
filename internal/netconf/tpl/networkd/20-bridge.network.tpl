{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.IfacesData*/ -}}
{{ .Comment }}
[Match]
Name=bridge

[Network]
{{- range .EVPNIfaces }}
VLAN=vlan{{ .VRF.ID }}
{{- end }}
{{- range .EVPNIfaces }}

[BridgeVLAN]
VLAN={{ .SVI.VLANID }}
{{- end }}