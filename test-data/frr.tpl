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
  # neighbor FABRIC route-map only-self-out out darf nicht aktiv sein! sonst problem!
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
{{- range $i, $ri := $v.RouteImports }}
  import vrf {{ $ri.SourceVRF }}
  import vrf route-map vrf{{ $v.ID }}-import-map
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
{{- range $i, $v := .VRFs }}
{{- range $j, $ri := $v.RouteImports }}
{{- range $k, $allowed := $ri.AllowedImportPrefixes }}
ip prefix-list vrf{{ $v.ID }}-import-prefixes seq 1{{ $j }}{{ $k }} permit {{ $allowed }}
{{- end }}
{{- end }}
!
route-map vrf{{ $v.ID }}-import-map permit 10
 match ip address prefix-list vrf{{ $v.ID }}-import-prefixes
!
route-map vrf{{ $v.ID }}-import-map deny 99
!
{{- end }}
line vty
!