package api

import (
	"os"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/crazyfrankie/goim/interfaces/user/api/handler"
	"github.com/crazyfrankie/goim/pkg/gin/middleware"
	userv1 "github.com/crazyfrankie/goim/protocol/user/v1"
)

// InitEngine returns gin.Engine.Currently, gRPC employs a direct connection approach,
// with plans to introduce a service registration and discovery mechanism in the future.
func InitEngine() (*gin.Engine, error) {
	srv := gin.Default()

	target := os.Getenv("SERVER_TARGET")
	cc, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	userCli := userv1.NewUserServiceClient(cc)
	userHdl := handler.NewUserHandler(userCli)

	authHdl, err := middleware.NewAuthnHandler(userCli)
	if err != nil {
		return nil, err
	}

	srv.Use(authHdl.IgnorePath([]string{"/api/user/login", "/api/user/register"}).Auth())

	apiGroup := srv.Group("api")
	userHdl.RegisterRoute(apiGroup)

	return srv, nil
}
