{{- /*gotype: git.f-i-ts.de/cloud-native/metal/metal-networker.FrrData*/ -}}
{{- $ASN := .ASN -}}
{{- $RouterId := .RouterID -}}
{{ .Comment }}
frr version {{ .FRRVersion }}
frr defaults datacenter
hostname {{ .Hostname }}
username cumulus nopassword
!
service integrated-vtysh-config
!
log syslog informational
{{ range .VRFs -}}
!
vrf vrf{{ .ID}}
 vni {{ .VNI }}
 {{ range .RouteLeaks -}}
 {{ . }}
 {{- end }}
{{ end -}}
!
interface eth0
 ipv6 nd ra-interval 6
 no ipv6 nd suppress-ra
!
interface eth1
 ipv6 nd ra-interval 6
 no ipv6 nd suppress-ra
!
router bgp {{ .ASN }}
 bgp router-id {{ .RouterID }}
 bgp bestpath as-path multipath-relax
 neighbor FABRIC peer-group
 neighbor FABRIC remote-as external
 neighbor FABRIC timers 1 3
 neighbor eth0 interface peer-group FABRIC
 neighbor eth1 interface peer-group FABRIC
 !
 address-family ipv4 unicast
  redistribute connected route-map LOOPBACKS
  neighbor FABRIC route-map only-self-out out
 exit-address-family
 !
 address-family l2vpn evpn
  neighbor FABRIC activate
  neighbor FABRIC route-map only-self-out out
  advertise-all-vni
 exit-address-family
!
{{ range $i, $v := .VRFs -}}
router bgp {{ $ASN }} vrf vrf{{ $v.ID }}
 bgp router-id {{ $RouterId }}
 bgp bestpath as-path multipath-relax
 !
 address-family ipv4 unicast
  redistribute connected
{{- range $v.NetworksAnnounced }}
  network {{ . }}
{{- end }}
 exit-address-family
 !
 address-family l2vpn evpn
  advertise ipv4 unicast
 exit-address-family
!
{{ end -}}
bgp as-path access-list SELF permit ^$
!
route-map only-self-out permit 10
 match as-path SELF
!
route-map only-self-out deny 99
!
route-map LOOPBACKS permit 10
 match interface lo
!
line vty
!