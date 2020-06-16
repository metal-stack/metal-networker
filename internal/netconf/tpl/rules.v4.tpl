{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.IptablesData*/ -}}
{{ .Comment }}
table ip metal {
    chain input {
        type filter hook input priority 0; policy drop;
        ct state established,related counter accept comment "stateful input"
        iifname "lo" counter accept comment "BGP unnumbered"
        iifname "lan0" ip saddr 10.0.0.0/8 udp dport 4789 counter accept comment "incoming VXLAN lan0"
        iifname "lan1" ip saddr 10.0.0.0/8 udp dport 4789 counter accept comment "incoming VXLAN lan1"
        tcp dport ssh ct state new counter accept comment "SSH incoming connections"
        ip saddr 10.0.0.0/8 tcp dport { 9000, 9100, 9630 } counter accept comment "firewall metrics"
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
        oifname "lo" counter accept comment "lo output required e.g. for chrony"
        ct state established,related counter accept comment "stateful output"
        ip daddr 10.0.0.0/8 udp dport 4789 counter accept comment "outgoing VXLAN"
        ct state invalid counter drop comment "drop invalid packets"
    }
    chain refuse {
        limit rate 2/minute counter log prefix "nftables-metal-dropped: "
        counter drop
    }
}
table ip nat {
    chain prerouting {
        type nat hook prerouting priority 0; policy accept;
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
        {{- range .SourceSpecs }}
        oifname "{{ $out }}" ip saddr {{ . }} counter masquerade comment "{{ $cmt }}"
        {{- end }}
        {{- end }}
    }
}