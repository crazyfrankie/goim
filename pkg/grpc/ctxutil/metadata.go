package ctxutil

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"
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
	userID, _ := strconv.ParseInt(uid[0], 10, 64)

	return userID
}

func MustGetUserAgent(ctx context.Context) string {
	md := GetMetadata(ctx)
	if md == nil {
		panic("mustGetUserAgentFromCtx: metadata is nil")
	}

	ua := md.Get("use_agent")

	return ua[0]
}
