{{- /*gotype: git.f-i-ts.de/cloud-native/metal/metal-networker/internal/netconf.NodeExporterData*/ -}}
{{ .Comment }}
[Unit]
Description=Node exporter - provides prometheus metrics about the node
After=network.target

[Service]
ExecStart=/usr/sbin/ip vrf exec {{ .TenantVrf }} /usr/local/bin/node_exporter --collector.tcpstat
Restart=always
RestartSec=30

[Install]
WantedBy=multi-user.target