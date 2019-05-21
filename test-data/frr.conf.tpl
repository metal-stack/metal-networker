{{ $root := . -}}
frr version 7.0
frr defaults datacenter
hostname {{ .Hostname }}
username cumulus nopassword
!
service integrated-vtysh-config
!
log syslog informational
{{ range .Networks }}
    {{ if .Primary }}
        {{- $Primary := . -}}
    {{ end }}
{{ end }}
!
# Primary Network
vrf vrf{{ $Primary.VRF }}
vni {{ .$Primary.VRF }}
{{ range .Networks }}
    {{ if .NetworkId == $Primary.NetworkId }}
{{ continue }}
  {{ end }}
  {{- $Net := . -}}
  {{ range .DestinationPrefixes }}
ip route {{ . }} vrf{{ $Net.Vrf }} nexthop-vrf vrf{{ $Net.Vrf }}
  {{ end }}
  {{ end }}
!
# External Networks
{{ range .Networks }}
  {{ if .NetworkId == $Primary.NetworkId }}
    {{ continue }}
  {{ end }}
vrf vrf{{ .Vrf }}
vni {{ .Vrf }}
ip route {{ $Primary.Network.Prefix }}/{{ $Primary.Network.Length }} vrf{{ $Primary.Vrf }} nexthop-vrf vrf{{ $Primary.Vrf }}
!
{{ end }}
interface eth1
ipv6 nd ra-interval 6
no ipv6 nd suppress-ra
!
interface eth2
ipv6 nd ra-interval 6
no ipv6 nd suppress-ra
!
router bgp {{ .ASN }}
bgp router-id {{ .IPAddress }}
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
# Primary Network BGP Instance
router bgp {{ $root.ASN }} vrf vrf{{ $Primary.Vrf }}
bgp router-id {{ $root.IPAddress }}
bgp bestpath as-path multipath-relax
address-family ipv4 unicast
redistribute connected
{{ range .Networks }}
  {{ if .NetworkId == $Primary.NetworkId }}
      {{ continue }}
  {{ end }}
  {{ range .DestinationPrefixes }}
network {{ . }}
  {{ end }}
  {{ end }}
exit-address-family
address-family l2vpn evpn
advertise ipv4 unicast
exit-address-family
!
# External Networks BGP Instances
{{ range .Networks }}
{{ if .NetworkId == $Primary.NetworkId }}
    {{ continue }}
{{ end }}
router bgp {{ $root.ASN }} vrf vrf{{ .Vrf }}
bgp router-id {{ $root.IPAddress }}
bgp bestpath as-path multipath-relax
!
address-family ipv4 unicast
redistribute connected
network {{ $Primary.Network.Prefix }}/{{ $Primary.Network.Length }}
exit-address-family
!
address-family l2vpn evpn
advertise ipv4 unicast
exit-address-family
!
{{ end }}
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