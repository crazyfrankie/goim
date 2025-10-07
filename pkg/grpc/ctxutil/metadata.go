package ctxutil

import (
	"context"

	"google.golang.org/grpc/metadata"

	"github.com/crazyfrankie/goim/pkg/lang/conv"
)

func GetMetadata(ctx context.Context) metadata.MD {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	return md
}

func MustGetUserIDFromCtx(ctx context.Context) int64 {
	md := GetMetadata(ctx)
	if md == nil {
		panic("mustGetUserIDFromCtx: metadata is nil")
	}

	uid := md.Get("user_id")
	userID, _ := conv.StrToInt64(uid[0])

	return userID
}
