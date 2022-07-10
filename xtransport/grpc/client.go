package grpc

import (
	"context"
	"google.golang.org/grpc"
	"net"
)

type Client struct {
	*grpc.ClientConn
	listener net.Listener
	ctx      context.Context
	err      error
	address  string
	opts     []grpc.DialOption
	unary    []grpc.UnaryClientInterceptor
	stream   []grpc.StreamClientInterceptor
}

func NewClient(opts ...ClientOption) *Client {
	cli := &Client{
		ctx:     context.TODO(),
		address: "0.0.0.0:0",
	}
	for _, o := range opts {
		o(cli)
	}

	var grpcOpts []grpc.DialOption
	if len(cli.unary) > 0 {
		grpcOpts = append(grpcOpts, grpc.WithChainUnaryInterceptor(cli.unary...))
	}
	if len(cli.stream) > 0 {
		grpcOpts = append(grpcOpts, grpc.WithChainStreamInterceptor(cli.stream...))
	}
	if len(cli.opts) > 0 {
		grpcOpts = append(grpcOpts, cli.opts...)
	}

	conn, err := grpc.DialContext(cli.ctx, cli.address, cli.opts...)
	cli.err = err
	cli.ClientConn = conn
	return cli
}

func (c *Client) Stop() error {
	return c.Close()
}

type ClientOption func(*Client)

func ClientOptions(opts ...grpc.DialOption) ClientOption {
	return func(client *Client) {
		client.opts = opts
	}
}

func WithAddr(addr string) ClientOption {
	return func(client *Client) {
		client.address = addr
	}
}

func WithContext(ctx context.Context) ClientOption {
	return func(client *Client) {
		client.ctx = ctx
	}
}
