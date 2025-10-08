package ws

import (
	"net/http"

	wsctx "github.com/crazyfrankie/goim/interfaces/ws/context"
	"github.com/crazyfrankie/goim/pkg/gin/response"
	"github.com/crazyfrankie/goim/pkg/logs"
	"github.com/crazyfrankie/goim/pkg/sonic"
)

func httpError(ctx *wsctx.Context, err error) {
	logs.CtxWarnf(ctx, "ws connection error, err: %v", err)
	httpJson(ctx.Writer, response.ParseError(err))
}

func httpJson(w http.ResponseWriter, data any) {
	body, err := sonic.Marshal(data)
	if err != nil {
		http.Error(w, "json marshal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
