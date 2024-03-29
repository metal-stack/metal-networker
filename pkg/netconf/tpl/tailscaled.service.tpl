[Unit]
Description=Tailscale node agent
Documentation=https://tailscale.com/kb/
After=network.target

[Service]
LimitMEMLOCK=infinity
User=root
Group=root
Type=notify
Environment="TS_NO_LOGS_NO_SUPPORT=true"
ExecStartPre=ip vrf exec {{ .DefaultRouteVrf }} /usr/local/bin/tailscaled --cleanup
ExecStart=/bin/ip vrf exec {{ .DefaultRouteVrf }} /usr/local/bin/tailscaled --port {{ .TailscaledPort }}
ExecStopPost=ip vrf exec {{ .DefaultRouteVrf }} /usr/local/bin/tailscaled --cleanup
Restart=on-failure

RuntimeDirectory=tailscale
RuntimeDirectoryMode=0755
StateDirectory=tailscale
StateDirectoryMode=0700
CacheDirectory=tailscale
CacheDirectoryMode=0750

[Install]
WantedBy=multi-user.target