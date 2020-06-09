{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.IptablesData*/ -}}
{{ .Comment }}
table ip6 metal {
    chain input {
        type filter hook input priority 0; policy drop;
        meta l4proto ipv6-icmp counter accept comment "icmpv6 input required for neighbor discovery"
        iifname "lo" counter accept comment "BGP unnumbered"
        ct state established,related counter accept comment "stateful input"
        iifname "lan0" ip6 saddr fe80::/64 tcp dport bgp counter accept comment "bgp unnumbered input from lan0"
        iifname "lan1" ip6 saddr fe80::/64 tcp dport bgp counter accept comment "bgp unnumbered input from lan1"
        ct state invalid counter drop comment "drop invalid packets to prevent malicious activity"
        counter jump refuse
    }
    chain forward {
        type filter hook forward priority 0; policy drop;
        ct state invalid counter drop comment "drop invalid packets from forwarding to prevent malicious activity"
        counter jump refuse
    }
    chain output {
        type filter hook output priority 0; policy drop;
        ct state established,related counter accept comment "stateful output"
        oifname "lo" counter accept comment "BGP unnumbered"
        meta l4proto ipv6-icmp counter accept comment "icmpv6 output required for neighbor discovery"
        oifname "lan0" ip6 saddr fe80::/64 tcp dport bgp counter accept comment "bgp unnumbered output at lan0"
        oifname "lan1" ip6 saddr fe80::/64 tcp dport bgp counter accept comment "bgp unnumbered output at lan1"
    }
    chain refuse {
        limit rate 2/minute counter log prefix "nftables-metal-dropped: "
        counter drop
    }
}