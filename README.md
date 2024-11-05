# metal-networker

Configures networking related resources such as interfaces, frr and nftables.

## Preconditions

Ubuntu/Debian operating system in place with the following packages installed:

- frr >= 10.0
- nftables

## Usage

metal-networker is used by the install-go binary as library in the metal-hammer. 
It is expected that the configuration file contains valid YAML. 
See [./internal/netconf/testdata/firewall.yaml](internal/netconf/testdata/firewall.yaml) for a valid configuration for firewalls
and [./internal/netconf/testdata/machine.yaml](internal/netconf/testdata/machine.yaml) for a valid configuration for machines.
