{{- /*gotype: git.f-i-ts.de/cloud-native/metal/metal-networker/internal/netconf.DroptailerData*/ -}}
{{ .Comment }}
[Unit]
Description=Droptailer
After=network.target

[Service]
Environment=DROPTAILER_SERVER_ADDRESS=droptailer:50051
Environment=DROPTAILER_PREFIXES_OF_DROPS="nftables-metal-dropped: ,nftables-firewall-dropped: "
Environment=DROPTAILER_CLIENT_CERTIFICATE=/etc/droptailer-client/droptailer-client.crt
Environment=DROPTAILER_CLIENT_KEY=/etc/droptailer-client/droptailer-client.key
ExecStart=/usr/bin/ip vrf exec {{ .TenantVrf }} /usr/local/bin/droptailer-client
Restart=always
RestartSec=10

[Install]
WantedBy=firewall-policy-controller.service