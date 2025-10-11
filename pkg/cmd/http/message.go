package http

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/crazyfrankie/goim/interfaces/http/message"
	"github.com/crazyfrankie/goim/pkg/cmd"
	"github.com/crazyfrankie/goim/pkg/gin/starthttp"
	"github.com/crazyfrankie/goim/pkg/lang/program"
	"github.com/crazyfrankie/goim/types/consts"
)

type MessageCmd struct {
	*cmd.RootCmd
}

func NewMessageCmd() *MessageCmd {
	msgCmd := &MessageCmd{
		RootCmd: cmd.NewRootCmd(program.GetProcessName(), consts.MessageApiName),
	}
	msgCmd.Command.RunE = func(cmd *cobra.Command, args []string) error {
		return msgCmd.runE()
	}

	return msgCmd
}

func (m *MessageCmd) Exec() error {
	return m.Execute()
}

func (m *MessageCmd) runE() error {
	listenAddr := os.Getenv("LISTEN_ADDR")

	return starthttp.Start(context.Background(), listenAddr, message.Start, time.Second*5)
}
