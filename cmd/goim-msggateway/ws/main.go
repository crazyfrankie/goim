package main

import (
	"github.com/crazyfrankie/goim/pkg/cmd/ws"
	"github.com/crazyfrankie/goim/pkg/lang/program"
)

func main() {
	if err := ws.NewWebsocketCmd().Exec(); err != nil {
		program.ExitWithError(err)
	}
}
