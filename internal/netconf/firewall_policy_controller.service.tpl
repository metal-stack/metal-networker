{{- /*gotype: github.com/metal-stack/metal-networker/internal/netconf.FirewallPolicyControllerData*/ -}}
{{ .Comment }}
[Unit]
Description=Firewall policy controller - generates nftable rules based on k8s resources
After=network.target

[Service]
Environment=FIREWALL_KUBECFG=/etc/firewall-policy-controller/.kubeconfig
ExecStart=/bin/ip vrf exec {{ .DefaultRouteVrf }} /usr/local/bin/firewall-policy-controller
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
