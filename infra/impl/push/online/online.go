package online

import (
	"os"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	"github.com/crazyfrankie/goim/infra/contract/push/online"
	"github.com/crazyfrankie/goim/infra/impl/push/online/allnode"
	"github.com/crazyfrankie/goim/pkg/lang/conv"
)

func NewOnlinePusher(conn discovery.Conn) online.OnlinePusher {
	maxConcurrentWorkers, _ := conv.StrToInt64(os.Getenv("ONLINE_MAX_WORKERS"))

	return allnode.NewDefaultAllNode(conn, maxConcurrentWorkers)
}
