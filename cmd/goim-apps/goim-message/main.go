package main

import (
	"github.com/crazyfrankie/goim/pkg/cmd/rpc"
	"github.com/crazyfrankie/goim/pkg/lang/program"
)

func main() {
	if err := rpc.NewMessageCmd().Exec(); err != nil {
		program.ExitWithError(err)
	}
}
