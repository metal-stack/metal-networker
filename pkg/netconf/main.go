package netconf

import "go.uber.org/zap"

var log *zap.SugaredLogger

//nolint
func init() {
	z, _ := zap.NewProduction()
	log = z.Sugar()
}
