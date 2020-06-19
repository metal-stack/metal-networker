[Unit]
Description=EveBox Agent

[Service]
LimitMEMLOCK=infinity
User=root
Group=root
Type=oneshot
ExecStart=/bin/ip vrf exec {{ .DefaultRouteVrf }} /usr/bin/evebox agent

[Install]
WantedBy=multi-user.target