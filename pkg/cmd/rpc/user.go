package rpc

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/apps/user"
	"github.com/crazyfrankie/goim/pkg/cmd"
	"github.com/crazyfrankie/goim/pkg/grpc/interceptor"
	"github.com/crazyfrankie/goim/pkg/grpc/startrpc"
	"github.com/crazyfrankie/goim/pkg/lang/program"
	"github.com/crazyfrankie/goim/types/consts"
)

type UserCmd struct {
	*cmd.RootCmd
}

func NewUserCmd() *UserCmd {
	userCmd := &UserCmd{
		RootCmd: cmd.NewRootCmd(program.GetProcessName(), consts.UserServiceName),
	}
	userCmd.Command.RunE = func(cmd *cobra.Command, args []string) error {
		return userCmd.runE()
	}

	return userCmd
}

func (u *UserCmd) Exec() error {
	return u.Execute()
}

func (u *UserCmd) runE() error {
	listenIP := os.Getenv("LISTEN_IP")
	registerIP := os.Getenv("REGISTER_IP")
	listenPort := os.Getenv("LISTEN_PORT")

	return startrpc.Start(context.Background(), listenIP, registerIP, listenPort, consts.UserServiceName, user.Start, userGrpcServerOption()...)
}

func userGrpcServerOption() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.CtxMDInterceptor(),
			interceptor.ResponseInterceptor(),
		),
	}
}
