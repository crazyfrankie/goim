package ws

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/crazyfrankie/goim/interfaces/ws"
	"github.com/crazyfrankie/goim/pkg/cmd"
	"github.com/crazyfrankie/goim/pkg/grpc/startrpc"
	"github.com/crazyfrankie/goim/pkg/lang/program"
	"github.com/crazyfrankie/goim/types/consts"
)

type WebsocketCmd struct {
	*cmd.RootCmd
}

func NewWebsocketCmd() *WebsocketCmd {
	wsCmd := &WebsocketCmd{
		RootCmd: cmd.NewRootCmd(program.GetProcessName(), consts.UserServiceName),
	}
	wsCmd.Command.RunE = func(cmd *cobra.Command, args []string) error {
		return wsCmd.runE()
	}

	return wsCmd
}

func (w *WebsocketCmd) Exec() error {
	return w.Execute()
}

func (w *WebsocketCmd) runE() error {
	return startrpc.Start(context.Background(), "", "", "", consts.UserServiceName, ws.Start)
}
