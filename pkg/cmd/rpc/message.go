package rpc

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/crazyfrankie/goim/apps/message"
	"github.com/crazyfrankie/goim/pkg/cmd"
	"github.com/crazyfrankie/goim/pkg/grpc/startrpc"
	"github.com/crazyfrankie/goim/pkg/lang/program"
)

const messageServiceName = "goim-rpc-message"

type MessageCmd struct {
	*cmd.RootCmd
}

func NewMessageCmd() *MessageCmd {
	messageCmd := &MessageCmd{
		RootCmd: cmd.NewRootCmd(program.GetProcessName(), authServiceName),
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

	return startrpc.Start(context.Background(), listenIP, registerIP, listenPort, messageServiceName, message.Start)
}
