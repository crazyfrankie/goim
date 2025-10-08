package user

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	"github.com/crazyfrankie/goim/interfaces/http/user/handler"
	"github.com/crazyfrankie/goim/pkg/gin/middleware"
	authv1 "github.com/crazyfrankie/goim/protocol/auth/v1"
	userv1 "github.com/crazyfrankie/goim/protocol/user/v1"
)

// Start returns gin.Engine.
func Start(ctx context.Context, client discovery.SvcDiscoveryRegistry) (http.Handler, error) {
	srv := gin.Default()

	userTarget := os.Getenv("USER_SERVER_TARGET")
	authTarget := os.Getenv("AUTH_SERVER_TARGET")
	userCC, err := client.GetConn(ctx, userTarget)
	if err != nil {
		return nil, err
	}
	authCC, err := client.GetConn(ctx, authTarget)
	if err != nil {
		return nil, err
	}
	userCli := userv1.NewUserServiceClient(userCC)
	authCli := authv1.NewAuthServiceClient(authCC)
	userHdl := handler.NewUserHandler(userCli)
	authHdl, err := middleware.NewAuthnHandler(authCli)
	if err != nil {
		return nil, err
	}

	srv.Use(authHdl.IgnorePath([]string{"/api/user/login", "/api/user/register"}).Auth())

	apiGroup := srv.Group("api")
	userHdl.RegisterRoute(apiGroup)

	return srv, nil
}
