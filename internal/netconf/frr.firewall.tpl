{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.FirewallFRRData*/ -}}
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
log syslog debugging
debug bgp updates
debug bgp nht
debug bgp update-groups
debug bgp zebra
{{ range .VRFs -}}
!
vrf vrf{{ .ID}}
 vni {{ .VNI }}
{{ end -}}
!
interface lan0
 ipv6 nd ra-interval 6
 no ipv6 nd suppress-ra
!
interface lan1
 ipv6 nd ra-interval 6
 no ipv6 nd suppress-ra
!
router bgp {{ .ASN }}
 bgp router-id {{ .RouterID }}
 bgp bestpath as-path multipath-relax
 neighbor FABRIC peer-group
 neighbor FABRIC remote-as external
 neighbor FABRIC timers 1 3
 neighbor lan0 interface peer-group FABRIC
 neighbor lan1 interface peer-group FABRIC
 !
 address-family ipv4 unicast
  redistribute connected route-map LOOPBACKS
  neighbor FABRIC route-map only-self-out out
 exit-address-family
 !
 address-family l2vpn evpn
  neighbor FABRIC activate
  advertise-all-vni
 exit-address-family
!
{{- range .VRFs }}
router bgp {{ $ASN }} vrf vrf{{ .ID }}
 bgp router-id {{ $RouterId }}
 bgp bestpath as-path multipath-relax
 !
 address-family ipv4 unicast
  redistribute connected
 {{- range .ImportVRFNames }}
  import vrf {{ . }}
 {{- end }}
  import vrf route-map vrf{{ .ID }}-import-map
 exit-address-family
 !
 address-family l2vpn evpn
  advertise ipv4 unicast
 exit-address-family
!
{{- end }}
{{- range .VRFs }}
 {{- range .IPPrefixLists }}
ip prefix-list {{ .Name }} {{ .Spec }}
 {{- end}}
 {{- range .RouteMaps }}
route-map {{ .Name }} {{ .Policy }} {{ .Order }}
  {{- range .Entries }}
 {{ . }}
  {{- end }}
 {{- end }}
!
{{- end }}
route-map only-self-out permit 10
 match as-path SELF
route-map only-self-out deny 20
!
route-map LOOPBACKS permit 10
 match interface lo
!
bgp as-path access-list SELF permit ^$
!
line vty
!