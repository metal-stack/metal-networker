[Unit]
Description=Tailscale client
After=tailscaled.service

[Service]
LimitMEMLOCK=infinity
User=root
Group=root
Type=notify
ExecStart=/bin/ip vrf exec {{ .DefaultRouteVrf }} /usr/local/bin/tailscale up --auth-key {{ .AuthKey }} --login-server {{ .Address }}
ExecStopPost=ip vrf exec {{ .DefaultRouteVrf }} /usr/local/bin/tailscale down
Restart=on-failure

[Install]
WantedBy=multi-user.target