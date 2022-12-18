package xgrpc

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	pbHealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"net"
	"sync"
	"time"
)

// grpc client

var (
	once     sync.Once
	grpcConn = make(map[string]*grpc.ClientConn)
)

type GRPCClientConf struct {
	Alias string
	Addr  string
}

func newGRPCClient(addr string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	retryOps := []grpc_retry.CallOption{
		grpc_retry.WithMax(3),
		grpc_retry.WithPerRetryTimeout(time.Second * 2),
		grpc_retry.WithBackoff(grpc_retry.BackoffLinearWithJitter(time.Second/2, 0.2)),
	}
	retry := grpc_retry.UnaryClientInterceptor(retryOps...)
	// lb: k8s headless svc(fmt.Sprintf("dns:///%s", addr))
	opts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(retry, UnaryClientTimeout(3*time.Second)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             time.Second,
			PermitWithoutStream: true,
		}),
		//grpc.WithBlock(),
	}
	c, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

func NewGRPCClient(conf []*GRPCClientConf) []*grpc.ClientConn {
	clients := make([]*grpc.ClientConn, 0, len(conf))
	once.Do(func() {
		for _, c := range conf {
			cli, err := newGRPCClient(c.Addr)
			if err != nil {
				panic(err)
			}
			grpcConn[c.Alias] = cli
			clients = append(clients, cli)
		}
	})
	return clients
}

func GetGRPCClient(alias string) *grpc.ClientConn {
	return grpcConn[alias]
}

func StopGRPCClient() {
	for _, c := range grpcConn {
		_ = c.Close()
	}
}

// grpc server

type GRPCServerConf struct {
	NetWork string
	Addr    string
}

type Server struct {
	Obj      interface{}
	Register func(s *grpc.Server, obj interface{})
}

type GRPCServer struct {
	*grpc.Server
	net.Listener
	healthServer *health.Server
}

func (s *Server) NewGRPCServer(conf *GRPCServerConf) (*GRPCServer, error) {
	entry := logrus.NewEntry(logrus.StandardLogger())
	opt := []grpc_logrus.Option{
		grpc_logrus.WithLevels(func(code codes.Code) logrus.Level {
			if code == codes.OK {
				return logrus.InfoLevel
			}
			return logrus.ErrorLevel
		}),
		grpc_logrus.WithMessageProducer(grpc_logrus.DefaultMessageProducer),
	}
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(entry, opt...),
			grpc_logrus.PayloadUnaryServerInterceptor(entry, func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return true }),
			grpc_validator.UnaryServerInterceptor(),
		), UnaryServerTimeout(3*time.Second)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second,
			MaxConnectionAge:      30 * time.Second,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}
	srv := grpc.NewServer(opts...)
	server := &GRPCServer{
		Server:       srv,
		healthServer: health.NewServer(),
	}
	pbHealth.RegisterHealthServer(srv, server.healthServer)
	s.Register(srv, s.Obj)
	listen, err := net.Listen(conf.NetWork, conf.Addr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	server.Listener = listen
	return server, nil
}

func (gs *GRPCServer) Start() {
	go func() {
		gs.healthServer.Resume()
		err := gs.Serve(gs.Listener)
		if err != nil {
			panic(errors.WithStack(err))
		}
	}()
}

func (gs *GRPCServer) Stop() {
	gs.healthServer.Shutdown()
	gs.GracefulStop()
}
