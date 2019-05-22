{{- /*gotype: git.f-i-ts.de/cloud-native/metal/metal-networker.InterfacesData*/ -}}
{{ .Comment }}

auto all

# {{ .Underlay.Comment }}
iface lo inet loopback
{{- range .Underlay.LoopbackIps }}
    address {{ . }}/32
{{- end }}

iface eth0 inet dhcp

iface eth1
    mtu 9216

iface eth2
    mtu 9216

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