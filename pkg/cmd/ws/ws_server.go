package ws

import (
	"context"
	"os"

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
		RootCmd: cmd.NewRootCmd(program.GetProcessName(), consts.MsgGatewayName),
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
	cfg := &startrpc.Config{
		ListenIP:        "",
		ListenPort:      "",
		RegisterIP:      "",
		RPCRegisterName: consts.MsgGatewayName,
		RPCServiceVer:   consts.MsgGatewayVer,
		MetricAddr:      os.Getenv("METRIC_ADDR"),
		CollectorAddr:   os.Getenv("COLLECTOR_ADDR"),
		RPCStart:        ws.Start,
		ServerOpts:      nil,
	}

	return startrpc.Start(context.Background(), cfg)
}
