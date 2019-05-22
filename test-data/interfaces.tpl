{{- /*gotype: git.f-i-ts.de/cloud-native/metal/metal-networker.InterfacesData*/ -}}
{{ .Comment }}

auto all

# {{ .Underlay.Comment }}
iface lo inet loopback
{{- range $i, $n := .Underlay.LoopbackIps }}
    address {{ $n }}/32
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

{{ range $i, $e := .EVPNInterfaces -}}
# {{ $e.SVI.Comment }}
iface vlan{{ $e.VRF.Id }}
    mtu 9000
    vlan-id {{ $e.SVI.VlanId }}
    vlan-raw-device bridge
    vrf vrf{{ $e.VRF.Id }}
    {{- range $j, $a := $e.SVI.Addresses }}
    address {{ $a }}/32
    {{- end }}

# {{ $e.VXLAN.Comment }}
iface vni{{ $e.VXLAN.Id }}
    mtu 9000
    bridge-access {{ $e.SVI.VlanId }}
    bridge-learning off
    mstpctl-bpduguard yes
    mstpctl-portbpdufilter yes
    vxlan-id {{ $e.VXLAN.Id }}
    vxlan-local-tunnelip {{ $e.VXLAN.TunnelIp }}

# {{ $e.VRF.Comment }}
iface vrf{{ $e.VRF.Id }}
    mtu 9000
    vrf-table auto

{{ end }}