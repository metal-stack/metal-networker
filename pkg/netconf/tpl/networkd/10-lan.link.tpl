{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.SystemdLinkData*/ -}}
{{ .Comment }}
[Match]
PermanentMACAddress={{ .MAC }}

[Link]
Name=lan{{ .Index }}
NamePolicy=
MTUBytes={{ .MTU }}