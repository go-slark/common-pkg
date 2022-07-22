package xgrpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

func UnaryClientTimeout(defaultTime time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var cancel context.CancelFunc
		if _, ok := ctx.Deadline(); !ok {
			defaultTimeout := defaultTime
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
			err = ctx.Err()
			if err == context.Canceled {
				err = status.Error(codes.Canceled, err.Error())
			} else if err == context.DeadlineExceeded {
				err = status.Error(codes.DeadlineExceeded, err.Error())
			}
			return nil, err
		}
	}
}
