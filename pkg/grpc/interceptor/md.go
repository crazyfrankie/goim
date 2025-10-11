package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/crazyfrankie/goim/pkg/ctxcache"
)

func CtxMDInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx = ctxcache.Init(ctx)

		md, _ := metadata.FromIncomingContext(ctx)

		for k, v := range md {
			ctxcache.Store(ctx, k, v)
		}

		return handler(ctx, req)
	}
}
