package xgrpc

import (
	"context"
	"google.golang.org/grpc"
	"sync"
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

func UnaryServerTimeout(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		var (
			resp interface{}
			err  error
			l    sync.Mutex
		)
		done := make(chan struct{})
		ch := make(chan interface{}, 1)
		go func() {
			defer func() {
				if p := recover(); p != nil {
					ch <- p
				}
			}()

			l.Lock()
			defer l.Unlock()
			resp, err = handler(ctx, req)
			close(done)
		}()

		select {
		case p := <-ch:
			panic(p)
		case <-done:
			l.Lock()
			defer l.Unlock()
			return resp, err
		case <-ctx.Done():
			l.Lock()
			defer l.Unlock()
			return nil, ctx.Err()
		}
	}
}