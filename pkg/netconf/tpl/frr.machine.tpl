{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.FirewallFRRData*/ -}}
{{- $ASN := .ASN -}}
{{- $RouterId := .RouterID -}}
{{ .Comment }}
frr version {{ .FRRVersion }}
frr defaults datacenter
hostname {{ .Hostname }}
username cumulus nopassword
allow-reserved-ranges
!
log syslog debugging
debug bgp updates
debug bgp nht
debug bgp update-groups
debug bgp zebra
!
interface lan0
 ipv6 nd ra-interval 6
 no ipv6 nd suppress-ra
!
interface lan1
 ipv6 nd ra-interval 6
 no ipv6 nd suppress-ra
!
no zebra nexthop kernel enable
!
router bgp {{ .ASN }}
 bgp router-id {{ .RouterID }}
 bgp bestpath as-path multipath-relax
 neighbor TOR peer-group
 neighbor TOR remote-as external
 neighbor TOR timers 2 8
 neighbor lan0 interface peer-group TOR
 neighbor lan1 interface peer-group TOR
 neighbor LOCAL peer-group
 neighbor LOCAL remote-as internal
 neighbor LOCAL timers 2 8
 neighbor LOCAL route-map local-in in
 bgp listen range 0.0.0.0/0 peer-group LOCAL
 !
 address-family ipv4 unicast
  redistribute connected
  redistribute kernel
  neighbor TOR route-map only-self-out out
 exit-address-family
 !
 address-family ipv6 unicast
  redistribute connected
  redistribute kernel
  neighbor TOR route-map only-self-out out
  neighbor TOR activate
 exit-address-family
!
bgp as-path access-list SELF permit ^$
!
route-map local-in permit 10
  set weight 32768
!
route-map only-self-out permit 10
 match as-path SELF
!
route-map only-self-out deny 99
!