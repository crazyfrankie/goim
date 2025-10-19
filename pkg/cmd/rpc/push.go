package rpc

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/apps/push"
	"github.com/crazyfrankie/goim/pkg/cmd"
	"github.com/crazyfrankie/goim/pkg/grpc/interceptor"
	"github.com/crazyfrankie/goim/pkg/grpc/startrpc"
	"github.com/crazyfrankie/goim/pkg/lang/program"
	"github.com/crazyfrankie/goim/types/consts"
)

type PushCmd struct {
	*cmd.RootCmd
}

func NewPushCmd() *PushCmd {
	pushCmd := &PushCmd{
		RootCmd: cmd.NewRootCmd(program.GetProcessName(), consts.UserServiceName),
	}
	pushCmd.Command.RunE = func(cmd *cobra.Command, args []string) error {
		return pushCmd.runE()
	}

	return pushCmd
}

func (p *PushCmd) Exec() error {
	return p.Execute()
}

func (p *PushCmd) runE() error {
	cfg := &startrpc.Config{
		ListenIP:        os.Getenv("LISTEN_IP"),
		ListenPort:      os.Getenv("LISTEN_PORT"),
		RegisterIP:      os.Getenv("REGISTER_IP"),
		RPCRegisterName: consts.PushServiceName,
		RPCServiceVer:   consts.PushServiceVer,
		MetricAddr:      os.Getenv("METRIC_ADDR"),
		CollectorAddr:   os.Getenv("COLLECTOR_ADDR"),
		RPCStart:        push.Start,
		ServerOpts:      pushGrpcServerOption(),
	}

	return startrpc.Start(context.Background(), cfg)
}

func pushGrpcServerOption() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptor.CtxMDInterceptor(),
			interceptor.ResponseInterceptor(),
		),
	}
}
