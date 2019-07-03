{{- /*gotype: git.f-i-ts.de/cloud-native/metal/metal-networker/internal/netconf.MachineIfacesData*/ -}}
{{ .CommonIfacesData.Comment }}
#
# See /etc/systemd/network for additional network configuration.

# {{ .CommonIfacesData.Underlay.Comment }}
auto lo
iface lo inet static
{{- range .CommonIfacesData.Underlay.LoopbackIps }}
    address {{ . }}/32
{{- end }}
