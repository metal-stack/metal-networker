{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.SystemdLinkConfig*/ -}}
{{ .Comment }}
[Match]
MACAddress={{ .MAC }}

[Link]
Name=lan{{ .Index }}
NamePolicy=mac
MTUBytes={{ .MTU }}