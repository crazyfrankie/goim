package rpc

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/apps/message"
	"github.com/crazyfrankie/goim/pkg/cmd"
	"github.com/crazyfrankie/goim/pkg/grpc/interceptor"
	"github.com/crazyfrankie/goim/pkg/grpc/startrpc"
	"github.com/crazyfrankie/goim/pkg/lang/program"
	"github.com/crazyfrankie/goim/types/consts"
)

type MessageCmd struct {
	*cmd.RootCmd
}

func NewMessageCmd() *MessageCmd {
	messageCmd := &MessageCmd{
		RootCmd: cmd.NewRootCmd(program.GetProcessName(), consts.MessageServiceName),
	}
	messageCmd.Command.RunE = func(cmd *cobra.Command, args []string) error {
		return messageCmd.runE()
	}

	return messageCmd
}

func (m *MessageCmd) Exec() error {
	return m.Execute()
}

func (m *MessageCmd) runE() error {
	cfg := &startrpc.Config{
		ListenIP:        os.Getenv("LISTEN_IP"),
		ListenPort:      os.Getenv("LISTEN_PORT"),
		RegisterIP:      os.Getenv("REGISTER_IP"),
		RPCRegisterName: consts.MessageServiceName,
		RPCServiceVer:   consts.MessageServiceVer,
		MetricAddr:      os.Getenv("METRIC_ADDR"),
		CollectorAddr:   os.Getenv("COLLECTOR_ADDR"),
		RPCStart:        message.Start,
		ServerOpts:      msgGrpcServerOption(),
	}

	return startrpc.Start(context.Background(), cfg)
}

func msgGrpcServerOption() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptor.CtxMDInterceptor(),
			interceptor.ResponseInterceptor(),
		),
	}
}
