package ws

import (
	"context"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	"github.com/crazyfrankie/goim/pkg/lang/conv"
)

func Start(ctx context.Context, client discovery.SvcDiscoveryRegistry, server grpc.ServiceRegistrar) error {
	wsPort := os.Getenv("LISTEN_PORT")
	wsMaxConnNum, _ := conv.StrToInt64(os.Getenv("WS_MAX_CONN"))
	wsTimeout, _ := conv.StrToInt64(os.Getenv("WS_TIMEOUT"))
	maxMsgLen, _ := conv.StrToInt64(os.Getenv("WS_MAX_MSG_LEN"))

	longServer := NewWebsocketServer(
		WithPort(wsPort),
		WithMaxConnNum(wsMaxConnNum),
		WithHandshakeTimeout(time.Duration(wsTimeout)*time.Second),
		WithMessageMaxMsgLength(int(maxMsgLen)),
	)

	if err := longServer.SetDiscoveryRegistry(ctx, client); err != nil {
		return err
	}

	go longServer.ChangeOnlineStatus(4)

	return longServer.Run(ctx)
}
