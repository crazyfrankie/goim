package message

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	"github.com/crazyfrankie/goim/interfaces/http/message/handler"
	"github.com/crazyfrankie/goim/pkg/gin/middleware"
	authv1 "github.com/crazyfrankie/goim/protocol/auth/v1"
	messagev1 "github.com/crazyfrankie/goim/protocol/message/v1"
	"github.com/crazyfrankie/goim/types/consts"
)

func Start(ctx context.Context, client discovery.SvcDiscoveryRegistry) (http.Handler, error) {
	srv := gin.Default()

	authCC, err := client.GetConn(ctx, consts.AuthServiceName)
	if err != nil {
		return nil, err
	}
	messageCC, err := client.GetConn(ctx, consts.MessageServiceName)
	if err != nil {
		return nil, err
	}

	messageCli := messagev1.NewMessageServiceClient(messageCC)
	authCli := authv1.NewAuthServiceClient(authCC)
	messageHdl := handler.NewMessageHandler(messageCli)
	authHdl, err := middleware.NewAuthnHandler(authCli)
	if err != nil {
		return nil, err
	}

	srv.Use(authHdl.Auth())

	apiGroup := srv.Group("api")
	messageHdl.RegisterRoute(apiGroup)

	return srv, nil
}
