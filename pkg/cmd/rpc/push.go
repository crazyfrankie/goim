package rpc

import (
	"context"
	"os"

	"github.com/spf13/cobra"
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
	listenIP := os.Getenv("LISTEN_IP")
	registerIP := os.Getenv("REGISTER_IP")
	listenPort := os.Getenv("LISTEN_PORT")
	metricAddr := os.Getenv("METRIC_ADDR")

	return startrpc.Start(context.Background(), listenIP, registerIP, listenPort, metricAddr, consts.PushServiceName, push.Start, pushGrpcServerOption()...)
}

func pushGrpcServerOption() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.CtxMDInterceptor(),
			interceptor.ResponseInterceptor(),
		),
	}
}
