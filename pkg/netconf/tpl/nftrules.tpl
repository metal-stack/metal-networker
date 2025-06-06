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

        {{ .DNSProxyDNAT.DestSpec.AddressFamily }} saddr {{ .DNSProxyDNAT.SAddr }} tcp dport {{ .DNSProxyDNAT.Port }} {{ .DNSProxyDNAT.DestSpec.AddressFamily }} daddr {{ .DNSProxyDNAT.DestSpec.Address }} accept comment "{{ .DNSProxyDNAT.Comment }}"
        {{ .DNSProxyDNAT.DestSpec.AddressFamily }} saddr {{ .DNSProxyDNAT.SAddr }} udp dport {{ .DNSProxyDNAT.Port }} {{ .DNSProxyDNAT.DestSpec.AddressFamily }} daddr {{ .DNSProxyDNAT.DestSpec.Address }} accept comment "{{ .DNSProxyDNAT.Comment }}"
        {{- end }}

        {{ if .VPN -}}
        iifname "tailscale*" accept comment "Accept tailscale traffic"
        {{- else -}}
        tcp dport ssh ct state new counter accept comment "SSH incoming connections"
        {{- end }}
        {{- range .Input.InInterfaces }}
        iifname "{{ . }}" tcp dport 9100 counter accept comment "node metrics"
        iifname "{{ . }}" tcp dport 9630 counter accept comment "nftables metrics"
        {{- end }}
        
        ct state invalid counter drop comment "drop invalid packets to prevent malicious activity"
        counter jump refuse
    }
    chain forward {
        type filter hook forward priority 0; policy {{ .ForwardPolicy }};
        ct state invalid counter drop comment "drop invalid packets from forwarding to prevent malicious activity"
        ct state established,related counter accept comment "stateful forward"
        tcp dport bgp ct state new counter jump refuse comment "block bgp forward to machines"
        {{- range .FirewallRules.Egress }}
        {{ . }}
        {{- end }}
        {{- range .FirewallRules.Ingress }}
        {{ . }}
        {{- end }}
        {{ if eq .ForwardPolicy "drop" -}}
        limit rate 2/minute counter log prefix "nftables-metal-dropped: "
        {{- end }}
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
    chain output_ct {
        type filter hook output priority raw; policy accept;
        {{-  $port:=.DNSProxyDNAT.Port }}
        {{-  $zone:=.DNSProxyDNAT.Zone }}
        {{- range .DNSProxyDNAT.InInterfaces }}
        oifname "{{ . }}" tcp sport {{ $port }} ct zone set {{ $zone }}
        oifname "{{ . }}" udp sport {{ $port }} ct zone set {{ $zone }}
        {{- end }}
    }
    chain refuse {
        limit rate 2/minute counter log prefix "nftables-metal-dropped: "
        counter drop
    }
}
table inet nat {
    set proxy_dns_servers {
    	type ipv4_addr
    	flags interval
    	auto-merge
    	elements = { 8.8.8.8, 8.8.4.4, 1.1.1.1, 1.0.0.1 }
    }
    {{- if eq .DNSProxyDNAT.DestSpec.AddressFamily "ip6" }}

    set proxy_dns_servers_v6 {
    	type ipv6_addr
    	flags interval
    	auto-merge
    	elements = { 2001:4860:4860::8888, 2001:4860:4860::8844, 2606:4700:4700::1111, 2606:4700:4700::1001 }
    }
    {{- end }}

    chain prerouting {
        type nat hook prerouting priority 0; policy accept;
        {{-  $port:=.DNSProxyDNAT.Port }}
        {{-  $dst:=.DNSProxyDNAT.DestSpec }}
        {{-  $daddr:=.DNSProxyDNAT.DAddr }}
        {{-  $cmt:=.DNSProxyDNAT.Comment }}
        {{- range .DNSProxyDNAT.InInterfaces }}
        {{ if $daddr -}} {{ $dst.AddressFamily }} daddr {{ $daddr }} {{ end -}} iifname "{{ . }}" tcp dport {{ $port }} dnat {{ $dst.AddressFamily }} to {{ $dst.Address }} comment "{{ $cmt }}"
        {{ if $daddr -}} {{ $dst.AddressFamily }} daddr {{ $daddr }} {{ end -}} iifname "{{ . }}" udp dport {{ $port }} dnat {{ $dst.AddressFamily }} to {{ $dst.Address }} comment "{{ $cmt }}"
        {{- end }}
    }
    chain prerouting_ct {
        type filter hook prerouting priority raw; policy accept;
        {{-  $port:=.DNSProxyDNAT.Port }}
        {{-  $zone:=.DNSProxyDNAT.Zone }}
        {{- range .DNSProxyDNAT.InInterfaces }}
        iifname "{{ . }}" tcp dport {{ $port }} ct zone set {{ $zone }}
        iifname "{{ . }}" udp dport {{ $port }} ct zone set {{ $zone }}
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
        oifname "{{ $out }}" {{ .AddressFamily }} saddr {{ .Address }} {{ .AddressFamily }} daddr != {{ $outspec.Address }} counter masquerade random comment "{{ $cmt }}"{{ else }}
        oifname "{{ $out }}" {{ .AddressFamily }} saddr {{ .Address }} counter masquerade random comment "{{ $cmt }}"
        {{- end }}
        {{- end }}
        {{- end }}
    }
}