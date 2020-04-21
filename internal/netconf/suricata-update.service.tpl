[Unit]
Description=Suricata Intrusion Detection Service Rules Update

[Service]
User=root
Group=root
Type=oneshot
ExecStart=/bin/ip vrf exec {{ .DefaultRouteVrf }} /usr/bin/suricata-update

[Install]
WantedBy=multi-user.target