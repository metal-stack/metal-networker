{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.NodeExporterData*/ -}}
{{ .Comment }}
[Unit]
Description=Node exporter - provides prometheus metrics about the node
After=network.target

[Service]
ExecStart=/bin/ip vrf exec {{ .TenantVrf }} /usr/local/bin/node_exporter --collector.tcpstat
Restart=always
RestartSec=30

[Install]
WantedBy=multi-user.target