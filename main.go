package main

import (
	"git.f-i-ts.de/cloud-native/metal/metal-networker/cmd"
	"git.f-i-ts.de/cloud-native/metallib/zapup"
)

var log = zapup.MustRootLogger().Sugar()

func main() {
	cmd.Execute()
}
