# metal-networker

Configures networking by applying setup from a configuration file.

## Configuration File

The configuration file is expected to contain YAML. The YAML structures define network interface configuration for  
ifup/ ifdown and Free Range Routing (FRR).

See `./test-data/install.yaml`

## Usage

The binary must be invoked with the configuration file as argument:

```bash
# metal-networker <config-file>
metal-networker install.yaml

```