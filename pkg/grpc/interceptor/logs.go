package interceptor

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/pkg/logs"
)

func ClientLogInterceptor() grpc.UnaryClientInterceptor {
	logTraceID := func(ctx context.Context) logging.Fields {
		if span := oteltrace.SpanContextFromContext(ctx); span.IsSampled() {
			return logging.Fields{"traceID", span.TraceID().String()}
		}
		return nil
	}

	return logging.UnaryClientInterceptor(initLog(), logging.WithFieldsFromContext(logTraceID))
}

func initLog() logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, level logging.Level, msg string, fields ...any) {
		f := make([]any, 0, len(fields)/2)

		msg = "[msg: " + msg + "] "

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			msg += " [" + key.(string) + ": %v]"
			f = append(f, value)
		}

		switch level {
		case logging.LevelDebug:
			logs.CtxDebugf(ctx, msg, f...)
		case logging.LevelInfo:
			logs.CtxInfof(ctx, msg, f...)
		case logging.LevelWarn:
			logs.CtxWarnf(ctx, msg, f...)
		case logging.LevelError:
			logs.CtxErrorf(ctx, msg, f...)
		default:
			panic(fmt.Sprintf("unknown level %v", level))
		}
	})
}
