package grpc

import "google.golang.org/grpc"

// GRPC Server

type RegisterObj struct {
	Obj      interface{}
	Register func(s *grpc.Server, obj interface{})
}

func (r *RegisterObj) NewGRPCServer() *Server {
	srv := NewServer()
	r.Register(srv.Server, r.Obj)
	return srv
}

// GRPC Client

type ClientGRPC struct {
	clients map[string]*Client
}

type ClientObj struct {
	Name string
	Addr string
}

func NewGRPCClient(objs []*ClientObj, opts ...ClientOption) *ClientGRPC {
	clients := make(map[string]*Client, len(objs))
	for _, obj := range objs {
		clients[obj.Name] = NewClient(append(append([]ClientOption{}, WithAddr(obj.Addr)), opts...)...)
	}
	return &ClientGRPC{clients: clients}
}

func (c *ClientGRPC) Stop() {
	for _, client := range c.clients {
		_ = client.Stop()
	}
}
