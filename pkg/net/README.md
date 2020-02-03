# network

Network can apply changes to `/etc/network/interfaces` and `/etc/frr/frr.conf`. 

It was intentionally created to provide a common means to:

- apply validation
- render interfaces/frr.conf files
- reload required services to apply changes

## Requirements

Network lib relies on `ifupdown2` and `systemd`. It also is assumed frr is installed as systemd service.

## Usage

Make use network lib:

```go
package main

import "github.com/metal-stack/metal-networker/pkg/net"

func main() {
	// TODO
}

```