package user

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/crazyfrankie/goim/interfaces/http/user/handler"
	"github.com/crazyfrankie/goim/pkg/gin/middleware"
	authv1 "github.com/crazyfrankie/goim/protocol/auth/v1"
	userv1 "github.com/crazyfrankie/goim/protocol/user/v1"
)

// Start returns gin.Engine.Currently, gRPC employs a direct connection approach,
// with plans to introduce a service registration and discovery mechanism in the future.
func Start() (http.Handler, error) {
	srv := gin.Default()

	userTarget := os.Getenv("USER_SERVER_TARGET")
	authTarget := os.Getenv("AUTH_SERVER_TARGET")
	userCC, err := grpc.NewClient(userTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	authCC, err := grpc.NewClient(authTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
