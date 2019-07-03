{{- /*gotype: git.f-i-ts.de/cloud-native/metal/metal-networker/internal/netconf.SystemdLinkConfig*/ -}}
{{ .Comment }}
[Match]
MACAddress={{ .MAC }}

[Link]
Name=lan{{ .Index }}
NamePolicy=mac
MTUBytes={{ .MTU }}