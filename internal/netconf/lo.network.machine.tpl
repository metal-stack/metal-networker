{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.MachineIfacesData*/ -}}
{{ .CommonIfacesData.Comment }}
#
# See /etc/systemd/network for additional network configuration.

# {{ .CommonIfacesData.Loopback.Comment }}
[Match]
Name=lo

[Address]
Address=127.0.0.1/8
{{- range .CommonIfacesData.Loopback.IPs }}
[Address]
Address={{ . }}/32
{{- end }}