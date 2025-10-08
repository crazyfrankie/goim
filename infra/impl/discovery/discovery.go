package discovery

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	"github.com/crazyfrankie/goim/infra/impl/discovery/etcd"
	"github.com/crazyfrankie/goim/types/consts"
)

func NewDiscoveryRegister() (discovery.SvcDiscoveryRegistry, error) {
	disTyp := os.Getenv(consts.DiscoveryType)

	switch disTyp {
	case "etcd":
		return initEtcdDis()
	default:
		return nil, fmt.Errorf("unsupported discovery type, %s", disTyp)
	}
}

func initEtcdDis() (discovery.SvcDiscoveryRegistry, error) {
	endpoint := os.Getenv("ETCD_ENDPOINT")
	endpoints := strings.Split(endpoint, ",")

	watchName := os.Getenv("WATCH_NAME")
	watchNames := strings.Split(watchName, ",")

	rootDir := os.Getenv("ROOT_DIR")

	userName := os.Getenv("ETCD_USER")
	password := os.Getenv("ETCD_PASSWORD")

	return etcd.NewSvcDiscoveryRegistry(rootDir, endpoints, watchNames,
		etcd.WithDialTimeout(10*time.Second),
		etcd.WithMaxCallSendMsgSize(20*1024*1024),
		etcd.WithAuth(userName, password))
}
