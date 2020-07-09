{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.FirewallControllerData*/ -}}
{{ .Comment }}
[Unit]
Description=Firewall controller - configures the firewall based on k8s resources
After=network.target

[Service]
LimitMEMLOCK=infinity
Environment=KUBECONFIG=/etc/firewall-controller/.kubeconfig
ExecStart=/bin/ip vrf exec {{ .DefaultRouteVrf }} /usr/local/bin/firewall-controller --service-ip {{ .ServiceIP }} --private-vrf {{ .PrivateVrfID }}
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
