package rpc

import (
	"context"
	"os"

	"github.com/spf13/cobra"
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
	listenIP := os.Getenv("LISTEN_IP")
	registerIP := os.Getenv("REGISTER_IP")
	listenPort := os.Getenv("LISTEN_PORT")

	return startrpc.Start(context.Background(), listenIP, registerIP, listenPort, consts.MessageServiceName, message.Start, msgGrpcServerOption()...)
}

func msgGrpcServerOption() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.CtxMDInterceptor(),
			interceptor.ResponseInterceptor(),
		),
	}
}
