{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.FirewallControllerData*/ -}}
{{ .Comment }}
[Unit]
Description=Firewall controller - configures the firewall based on k8s resources
After=network.target

[Service]
LimitMEMLOCK=infinity
# FIXME this path is comming from gardener-extension-provider-metal, should be changed to something less specific
# Also binary location is not changed for now
Environment=KUBECONFIG=/etc/firewall-controller/.kubeconfig
ExecStart=/bin/ip vrf exec {{ .DefaultRouteVrf }} /usr/local/bin/firewall-controller
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
