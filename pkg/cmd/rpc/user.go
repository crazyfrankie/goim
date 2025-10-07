package rpc

import (
	"context"
	"os"

	"github.com/crazyfrankie/goim/apps/user"
	"github.com/crazyfrankie/goim/pkg/lang/program"
	"github.com/spf13/cobra"

	"github.com/crazyfrankie/goim/pkg/cmd"
	"github.com/crazyfrankie/goim/pkg/grpc/startrpc"
)

const userServiceName = "goim-rpc-user"

type UserCmd struct {
	*cmd.RootCmd
}

func NewUserCmd() *UserCmd {
	userCmd := &UserCmd{
		RootCmd: cmd.NewRootCmd(program.GetProcessName(), userServiceName),
	}
	userCmd.Command.RunE = func(cmd *cobra.Command, args []string) error {
		return userCmd.runE()
	}

	return userCmd
}

func (u *UserCmd) Exec() error {
	return u.Execute()
}

func (u *UserCmd) runE() error {
	listenIP := os.Getenv("LISTEN_IP")
	registerIP := os.Getenv("REGISTER_IP")
	listenPort := os.Getenv("LISTEN_PORT")

	return startrpc.Start(context.Background(), listenIP, registerIP, listenPort, userServiceName, user.Start)
}
