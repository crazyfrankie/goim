package http

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/crazyfrankie/goim/interfaces/http/user"
	"github.com/crazyfrankie/goim/pkg/cmd"
	"github.com/crazyfrankie/goim/pkg/gin/starthttp"
	"github.com/crazyfrankie/goim/pkg/lang/program"
	"github.com/crazyfrankie/goim/types/consts"
)

type UserCmd struct {
	*cmd.RootCmd
}

func NewUserCmd() *UserCmd {
	userCmd := &UserCmd{
		RootCmd: cmd.NewRootCmd(program.GetProcessName(), consts.UserApiName),
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
	cfg := &starthttp.Config{
		ListenAddr:      os.Getenv("LISTEN_ADDR"),
		ServiceName:     consts.UserApiName,
		ServiceVer:      consts.UserApiVer,
		ShutdownTimeout: time.Second * 5,
		MetricAddr:      os.Getenv("METRIC_ADDR"),
		CollectorAddr:   os.Getenv("COLLECTOR_ADDR"),
		InitFunc:        user.Start,
	}

	return starthttp.Start(context.Background(), cfg)
}
