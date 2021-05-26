{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.IptablesData*/ -}}
{{ .Comment }}
table inet metal {
    chain input {
        type filter hook input priority 0; policy drop;
        meta l4proto ipv6-icmp counter accept comment "icmpv6 input required for neighbor discovery"
        iifname "lo" counter accept comment "BGP unnumbered"
        iifname "lan0" ip6 saddr fe80::/64 tcp dport bgp counter accept comment "bgp unnumbered input from lan0"
        iifname "lan1" ip6 saddr fe80::/64 tcp dport bgp counter accept comment "bgp unnumbered input from lan1"
        iifname "lan0" ip saddr 10.0.0.0/8 udp dport 4789 counter accept comment "incoming VXLAN lan0"
        iifname "lan1" ip saddr 10.0.0.0/8 udp dport 4789 counter accept comment "incoming VXLAN lan1"

        ct state established,related counter accept comment "stateful input"
        {{- if .DNSProxyDNAT.DestSpec.Address }}

        tcp dport {{ .DNSProxyDNAT.Port }} {{ .DNSProxyDNAT.DestSpec.AddressFamily }} daddr {{ .DNSProxyDNAT.DestSpec.Address }} accept comment "{{ .DNSProxyDNAT.Comment }}"
        udp dport {{ .DNSProxyDNAT.Port }} {{ .DNSProxyDNAT.DestSpec.AddressFamily }} daddr {{ .DNSProxyDNAT.DestSpec.Address }} accept comment "{{ .DNSProxyDNAT.Comment }}"
        {{- end }}

        tcp dport ssh ct state new counter accept comment "SSH incoming connections"
        ip saddr 10.0.0.0/8 tcp dport 9100 counter accept comment "node metrics"
        ip saddr 10.0.0.0/8 tcp dport 9630 counter accept comment "nftables metrics"
        
        ct state invalid counter drop comment "drop invalid packets to prevent malicious activity"
        counter jump refuse
    }
    chain forward {
        type filter hook forward priority 0; policy accept;
        ct state invalid counter drop comment "drop invalid packets from forwarding to prevent malicious activity"
        tcp dport bgp ct state new counter jump refuse comment "block bgp forward to machines"
    }
    chain output {
        type filter hook output priority 0; policy accept;
        meta l4proto ipv6-icmp counter accept comment "icmpv6 output required for neighbor discovery"
        oifname "lo" counter accept comment "lo output required e.g. for chrony"
        oifname "lan0" ip6 saddr fe80::/64 tcp dport bgp counter accept comment "bgp unnumbered output at lan0"
        oifname "lan1" ip6 saddr fe80::/64 tcp dport bgp counter accept comment "bgp unnumbered output at lan1"

        ip daddr 10.0.0.0/8 udp dport 4789 counter accept comment "outgoing VXLAN"
        
        ct state established,related counter accept comment "stateful output"
        ct state invalid counter drop comment "drop invalid packets"                
    }
    chain refuse {
        limit rate 2/minute counter log prefix "nftables-metal-dropped: "
        counter drop
    }
}
table inet nat {
    chain prerouting {
        type nat hook prerouting priority 0; policy accept;
        {{-  $port:=.DNSProxyDNAT.Port }}
        {{-  $dst:=.DNSProxyDNAT.DestSpec }}
        {{-  $cmt:=.DNSProxyDNAT.Comment }}
        {{- range .DNSProxyDNAT.InInterfaces }}
        iifname "{{ . }}" tcp dport {{ $port }} dnat {{ $dst.AddressFamily }} to {{ $dst.Address }} comment "{{ $cmt }}"
        iifname "{{ . }}" udp dport {{ $port }} dnat {{ $dst.AddressFamily }} to {{ $dst.Address }} comment "{{ $cmt }}"
        {{- end }}
    }
    chain input {
        type nat hook input priority 0; policy accept;
    }
    chain output {
        type nat hook output priority 0; policy accept;
    }
    chain postrouting {
        type nat hook postrouting priority 0; policy accept;
        {{- range .SNAT }}
        {{- $cmt:=.Comment }}
        {{- $out:=.OutInterface }}
        {{- $outspec:=.OutIntSpec }}
        {{- range .SourceSpecs }}
        {{- if and $outspec.Address (eq $outspec.AddressFamily .AddressFamily) }}
        oifname "{{ $out }}" {{ .AddressFamily }} saddr {{ .Address }} {{ .AddressFamily }} daddr != {{ $outspec.Address }} counter masquerade comment "{{ $cmt }}"{{ else }}
        oifname "{{ $out }}" {{ .AddressFamily }} saddr {{ .Address }} counter masquerade comment "{{ $cmt }}"
        {{- end }}
        {{- end }}
        {{- end }}
    }
}