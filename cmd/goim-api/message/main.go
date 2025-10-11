package main

import (
	"github.com/crazyfrankie/goim/pkg/cmd/http"
	"github.com/crazyfrankie/goim/pkg/lang/program"
)

func main() {
	if err := http.NewMessageCmd().Exec(); err != nil {
		program.ExitWithError(err)
	}
}
