package netconf

import "go.uber.org/zap"

func init() {
	log = zap.NewNop().Sugar()
}
