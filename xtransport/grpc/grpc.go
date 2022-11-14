package grpc

import (
	"github.com/smallfish-root/common-pkg/xencoding"
	"google.golang.org/grpc"
	"os"
)

// GRPC Server

type RegisterObj struct {
	Obj      interface{}
	Register func(s *grpc.Server, obj interface{})
}

func (r *RegisterObj) NewGRPCServer(opts ...ServerOption) *Server {
	srv := NewServer(opts...)
	r.Register(srv.Server, r.Obj)
	return srv
}

// GRPC Client

type GRPCClient struct {
	clients map[string]*Client
}

type ClientObj struct {
	Name string
	Addr string
}

func NewGRPCClient(objs []*ClientObj, opts []ClientOption) *GRPCClient {
	clients := make(map[string]*Client, len(objs))
	optsNum := len(opts)
	for _, obj := range objs {
		dstOpts := make([]ClientOption, 0, optsNum)
		if optsNum != 0 {
			err := xencoding.DeepCopy(dstOpts, opts)
			if err != nil {
				panic(err)
			}
		}
		client := NewClient(append(append([]ClientOption{}, WithAddr(obj.Addr)), dstOpts...)...)
		if client.err != nil {
			os.Exit(800)
		}
		clients[obj.Name] = client
	}
	return &GRPCClient{clients: clients}
}

func (c *GRPCClient) GetGRPCClient(name string) *Client {
	return c.clients[name]
}

func (c *GRPCClient) Stop() {
	for _, client := range c.clients {
		_ = client.Stop()
	}
}
