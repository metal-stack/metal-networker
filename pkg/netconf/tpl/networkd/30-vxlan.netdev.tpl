{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.EVPNIface*/ -}}
{{ .VXLAN.Comment }}
[NetDev]
Name=vni{{ .VXLAN.ID }}
Kind=vxlan

[VXLAN]
VNI={{ .VXLAN.ID }}
Local={{ .VXLAN.TunnelIP }}
UDPChecksum=true
MacLearning=false
DestinationPort=4789
