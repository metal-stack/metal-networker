# metal-networker

Configures networking related resources such as interfaces, frr, iptables.

## Preconditions

Ubuntu operating system in place with the following packages installed: 

- ifupdown2
- frr > 7.0
- iptables
- iptables-persistence

## Download

metal-networker is available from the blobstore:
 
 - [latest](https://blobstore.fi-ts.io/cloud-native/metal-networker-latest.tar.gz): Bleeding edge. Contains bugs and issues.
 - [stable](https://blobstore.fi-ts.io/cloud-native/metal-networker-stable.tar.gz): Contains no known issues. Considered ready for use in production.

## Usage

metal-networker must be invoked with the configuration file as argument. It is expected that the configuration file 
contains valid YAML. See [./internal/netconf/testdata/firewall.yaml](internal/netconf/testdata/firewall.yaml) for a valid configuration for firewalls and [./internal/netconf/testdata/machine.yaml](internal/netconf/testdata/machine.yaml) for a valid configuration for machines.

```bash
# metal-networker <config-file>
./metal-networker machine|firewall configure --input install.yaml

```
