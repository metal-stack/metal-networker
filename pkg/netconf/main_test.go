package netconf

import "go.uber.org/zap"

//nolint
func init() {
	log = zap.NewNop().Sugar()
}
