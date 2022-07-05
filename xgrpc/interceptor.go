package xgrpc

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

func UnaryClientTimeout() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var cancel context.CancelFunc
		if _, ok := ctx.Deadline(); !ok {
			defaultTimeout := 3 * time.Second
			ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		}
		if cancel != nil {
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
