[Unit]
Description=Suricata Intrusion Detection Service
After=network.target

[Service]
EnvironmentFile=-/etc/default/suricata
ExecStartPre=/bin/rm -f $PIDFILE
{{- if .EnableIPS }}
ExecStart=/usr/bin/suricata -c $SURCONF --pidfile $PIDFILE -q 0
{{- else }}
ExecStart=/usr/bin/suricata -c $SURCONF --pidfile $PIDFILE -i $IFACE
{{- end }}
ExecReload=/bin/kill -USR2 $MAINPID
Restart=always
RestartSec=60

[Install]
WantedBy=multi-user.target