package main

import (
	"github.com/metal-stack/metal-networker/cmd"
	"go.uber.org/zap"
)

func main() {
	z, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	logger := z.Sugar()
	cmd.Execute(logger)
}
