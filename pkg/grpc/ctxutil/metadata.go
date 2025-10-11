package ctxutil

import (
	"context"

	"github.com/crazyfrankie/goim/pkg/ctxcache"
	"github.com/crazyfrankie/goim/pkg/lang/conv"
)

func MustGetUserIDFromCtx(ctx context.Context) int64 {
	val, ok := ctxcache.Get[[]string](ctx, "user_id")
	if !ok {
		panic("mustGetUserIDFromCtx: metadata is nil")
	}

	userID, _ := conv.StrToInt64(val[0])

	return userID
}
