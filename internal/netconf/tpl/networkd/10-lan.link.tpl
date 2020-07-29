{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.SystemdLinkData*/ -}}
{{ .Comment }}
[Match]
MACAddress={{ .MAC }}

[Link]
Name=lan{{ .Index }}
NamePolicy=mac
MTUBytes={{ .MTU }}