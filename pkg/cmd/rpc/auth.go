package rpc

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/apps/auth"
	"github.com/crazyfrankie/goim/pkg/cmd"
	"github.com/crazyfrankie/goim/pkg/grpc/interceptor"
	"github.com/crazyfrankie/goim/pkg/grpc/startrpc"
	"github.com/crazyfrankie/goim/pkg/lang/program"
	"github.com/crazyfrankie/goim/types/consts"
)

type AuthCmd struct {
	*cmd.RootCmd
}

func NewAuthCmd() *AuthCmd {
	authCmd := &AuthCmd{
		RootCmd: cmd.NewRootCmd(program.GetProcessName(), consts.AuthServiceName),
	}
	authCmd.Command.RunE = func(cmd *cobra.Command, args []string) error {
		return authCmd.runE()
	}

	return authCmd
}

func (a *AuthCmd) Exec() error {
	return a.Execute()
}

func (a *AuthCmd) runE() error {
	listenIP := os.Getenv("LISTEN_IP")
	registerIP := os.Getenv("REGISTER_IP")
	listenPort := os.Getenv("LISTEN_PORT")

	return startrpc.Start(context.Background(), listenIP, registerIP, listenPort, consts.AuthServiceName, auth.Start, authGrpcServerOption()...)
}

func authGrpcServerOption() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.CtxMDInterceptor(),
			interceptor.ResponseInterceptor(),
		),
	}
}
