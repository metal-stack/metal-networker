# metal-networker

Configures ifup/ifdown and Free Range Routing (FRR) by applying setup from a configuration file.

## Configuration File

The configuration file is expected to contain YAML. See [./test-data/install.yaml](./test-data/install.yaml).

## Usage

The binary is invoked with the configuration file as argument:

```bash
# metal-networker <config-file>
metal-networker install.yaml

```