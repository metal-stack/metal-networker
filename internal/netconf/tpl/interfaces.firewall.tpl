{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.FirewallIfacesData*/ -}}
{{ .Comment }}
#
# See /etc/systemd/network for additional network configuration.

auto all

# {{ .Loopback.Comment }}
iface lo inet loopback
{{- range .Loopback.IPs }}
    address {{ . }}/32
{{- end }}

iface bridge
    bridge-ports {{ .Bridge.Ports }}
    bridge-vids {{ .Bridge.Vids }}
    bridge-vlan-aware yes

{{ range .EVPNInterfaces -}}
# {{ .SVI.Comment }}
iface vlan{{ .VRF.ID }}
    mtu 9000
    vlan-id {{ .SVI.VlanID }}
    vlan-raw-device bridge
    vrf vrf{{ .VRF.ID }}
    {{- range .SVI.Addresses }}
    address {{ . }}/32
    {{- end }}

# {{ .VXLAN.Comment }}
iface vni{{ .VXLAN.ID }}
    mtu 9000
    bridge-access {{ .SVI.VlanID }}
    bridge-learning off
    mstpctl-bpduguard yes
    mstpctl-portbpdufilter yes
    vxlan-id {{ .VXLAN.ID }}
    vxlan-local-tunnelip {{ .VXLAN.TunnelIP }}

# {{ .VRF.Comment }}
iface vrf{{ .VRF.ID }}
    mtu 9000
    vrf-table auto

{{ end }}