package rpc

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
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
	cfg := &startrpc.Config{
		ListenIP:        os.Getenv("LISTEN_IP"),
		ListenPort:      os.Getenv("LISTEN_PORT"),
		RegisterIP:      os.Getenv("REGISTER_IP"),
		RPCRegisterName: consts.AuthServiceName,
		RPCServiceVer:   consts.AuthServiceVer,
		MetricAddr:      os.Getenv("METRIC_ADDR"),
		CollectorAddr:   os.Getenv("COLLECTOR_ADDR"),
		RPCStart:        auth.Start,
		ServerOpts:      authGrpcServerOption(),
	}

	return startrpc.Start(context.Background(), cfg)
}

func authGrpcServerOption() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptor.CtxMDInterceptor(),
			interceptor.ResponseInterceptor(),
		),
	}
}
