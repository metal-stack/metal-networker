{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.NftablesExporterData*/ -}}
{{ .Comment }}
[Unit]
Description=Nftables exporter - provides prometheus metrics for nftables
After=network.target

[Service]
ExecStart=/usr/bin/nftables-exporter --config=/etc/nftables_exporter.yaml
Restart=always
RestartSec=30

[Install]
WantedBy=multi-user.target