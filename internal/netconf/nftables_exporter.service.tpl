{{- /*gotype: git.f-i-ts.de/cloud-native/metal/metal-networker/internal/netconf.NftablesExporterData*/ -}}
{{ .Comment }}
[Unit]
Description=Nftables exporter - provides prometheus metrics for nftables
After=network.target

[Service]
ExecStart=/bin/ip vrf exec {{ .TenantVrf }} /usr/local/bin/nftables_exporter --config=/etc/nftables_exporter.yaml
Restart=always
RestartSec=30

[Install]
WantedBy=multi-user.target