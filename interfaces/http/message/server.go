package message

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	"github.com/crazyfrankie/goim/interfaces/http/message/handler"
	"github.com/crazyfrankie/goim/pkg/gin/middleware"
	authv1 "github.com/crazyfrankie/goim/protocol/auth/v1"
	msgv1 "github.com/crazyfrankie/goim/protocol/msg/v1"
	"github.com/crazyfrankie/goim/types/consts"
)

func Start(ctx context.Context, client discovery.SvcDiscoveryRegistry, middlewares ...gin.HandlerFunc) (http.Handler, error) {
	srv := gin.Default()

	authCC, err := client.GetConn(ctx, consts.AuthServiceName)
	if err != nil {
		return nil, err
	}
	messageCC, err := client.GetConn(ctx, consts.MessageServiceName)
	if err != nil {
		return nil, err
	}

	messageCli := msgv1.NewMessageServiceClient(messageCC)
	authCli := authv1.NewAuthServiceClient(authCC)
	messageHdl := handler.NewMessageHandler(messageCli)
	authHdl, err := middleware.NewAuthnHandler(authCli)
	if err != nil {
		return nil, err
	}

	middlewares = append(middlewares, authHdl.Auth())

	srv.Use(middlewares...)

	apiGroup := srv.Group("api")
	messageHdl.RegisterRoute(apiGroup)

	return srv, nil
}
