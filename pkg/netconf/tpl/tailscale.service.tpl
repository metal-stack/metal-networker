[Unit]
Description=Tailscale client
After=tailscaled.service

[Service]
LimitMEMLOCK=infinity
User=root
Group=root
ExecStart=/bin/ip vrf exec {{ .DefaultRouteVrf }} /usr/local/bin/tailscale up --auth-key {{ .AuthKey }} --login-server {{ .Address }}
Restart=on-failure

[Install]
WantedBy=multi-user.target